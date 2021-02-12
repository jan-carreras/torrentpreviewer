package bittorrentproto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"io/ioutil"
	"prevtorrent/internal/preview"
	"time"
)

type TorrentClient struct {
	client *torrent.Client
}

func NewTorrentClient(client *torrent.Client) *TorrentClient {
	return &TorrentClient{client: client}
}

func (r *TorrentClient) Resolve(ctx context.Context, m preview.Magnet) ([]byte, error) {
	t, err := r.client.AddMagnet(m.Value())
	if err != nil {
		return nil, err
	}

	if err := r.waitForInfo(ctx, t); err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = t.Metainfo().Write(buf)
	return buf.Bytes(), err
}

func (r *TorrentClient) waitForInfo(ctx context.Context, t *torrent.Torrent) error {
	select {
	case <-t.GotInfo():
		return nil
	case <-ctx.Done():
		return errors.New("context cancelled while trying to get info")
	}
}

// TODO: Maybe pass the torrent ID would be better instead of passwing a DownloadPlan with a raw... IDK
func (r *TorrentClient) DownloadParts(ctx context.Context, downloadPlan preview.DownloadPlan) ([]preview.DownloadedPart, error) {
	t, err := r.getTorrent(downloadPlan)
	if err != nil {
		return nil, err
	}

	waitingFor := countNumberPiecesWaitingFor(t, downloadPlan)
	downloadPieces(t, downloadPlan)
	waitPiecesToDownload(ctx, t, waitingFor)

	fmt.Println("We have all the pieces, we can start bundling them together as a response!")

	downloads, err := bundleResponses(t, downloadPlan)

	if err != nil {
		return nil, err
	}
	fmt.Println("downloads=", len(downloads))
	for i, dw := range downloads {
		fmt.Println(i, len(dw.Data()), dw.PieceRange())
		err = ioutil.WriteFile("/tmp/it-works.mp4", dw.Data(), 0666)
		if err != nil {
			return nil, err
		}

	}
	return downloads, nil
}

func bundleResponses(t *torrent.Torrent, downloadPlan preview.DownloadPlan) ([]preview.DownloadedPart, error) {
	downloads := make([]preview.DownloadedPart, 0)
	for _, plan := range downloadPlan.GetPlan() {
		piece := new(bytes.Buffer)
		for pieceIdx := plan.Start(); pieceIdx <= plan.End(); pieceIdx++ {
			buf := make([]byte, plan.EndOffset(pieceIdx))
			off := plan.StartOffset(pieceIdx)
			n, err := t.Piece(pieceIdx).Storage().ReadAt(buf, int64(off))
			if n != len(buf) {
				return nil, errors.New("not reading all the piece")
			}
			if err != nil {
				return nil, err
			}
			_, err = piece.Write(buf)
			if err != nil {
				return nil, err
			}
		}
		download := preview.NewDownloadedPart(plan, piece.Bytes())
		downloads = append(downloads, download)
	}

	return downloads, nil
}

func (r *TorrentClient) getTorrent(plan preview.DownloadPlan) (*torrent.Torrent, error) {
	buff := bytes.NewBuffer(plan.GetTorrent().Raw())
	metaInfo, err := metainfo.Load(buff)
	if err != nil {
		return nil, err
	}
	return r.client.AddTorrent(metaInfo)
}

func countNumberPiecesWaitingFor(t *torrent.Torrent, downloadPlan preview.DownloadPlan) int {
	waitingFor := 0
	for _, plan := range downloadPlan.GetPlan() {
		for pIdx := plan.Start(); pIdx <= plan.End(); pIdx++ {
			if t.Piece(pIdx).State().Complete {
				continue
			}
			waitingFor++
		}
	}
	return waitingFor
}

func downloadPieces(t *torrent.Torrent, downloadPlan preview.DownloadPlan) {
	// Idempotent. All the pieces already downloaded are ignored.
	for _, plan := range downloadPlan.GetPlan() {
		t.DownloadPieces(plan.Start(), plan.End()+1)
	}
}

func waitPiecesToDownload(ctx context.Context, t *torrent.Torrent, waitingFor int) {
	for waitingFor > 0 {
		fmt.Println("number of connected peers:", len(t.PeerConns())) // TODO: To logger
		select {
		case _v := <-t.SubscribePieceStateChanges().Values:
			v, ok := _v.(torrent.PieceStateChange)
			if !ok {
				break
			}
			if v.Complete {
				fmt.Println(v.Index, "complete:", v.Complete)
				waitingFor--
			}
		case <-time.After(time.Second * 3): // TODO: This has to go away, eventually
		case <-ctx.Done():
			break
		}
	}
}
