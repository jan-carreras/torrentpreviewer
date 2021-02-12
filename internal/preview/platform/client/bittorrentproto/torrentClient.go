package bittorrentproto

import (
	"bytes"
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/preview"
	"time"
)

type TorrentClient struct {
	client *torrent.Client
	logger *logrus.Logger
}

func NewTorrentClient(client *torrent.Client, logger *logrus.Logger) *TorrentClient {
	return &TorrentClient{client: client, logger: logger}
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

func (r *TorrentClient) DownloadParts(ctx context.Context, downloadPlan preview.DownloadPlan) ([]preview.DownloadedPart, error) {
	t, err := r.getTorrent(downloadPlan)
	if err != nil {
		return nil, err
	}

	waitingFor := countNumberPiecesWaitingFor(t, downloadPlan)
	downloadPieces(t, downloadPlan)
	r.waitPiecesToDownload(ctx, t, waitingFor)

	r.logger.WithFields(logrus.Fields{
		"torrent":         t.Name(),
		"partsWaitingFor": waitingFor,
	}).Info("we have all the pieces, we can start bundling them together as a response")

	downloads, err := bundleResponses(t, downloadPlan)

	if err != nil {
		return nil, err
	}
	r.logger.WithFields(logrus.Fields{
		"torrent":        t.Name(),
		"downloadsCount": len(downloads),
	}).Info("number of downloaded files")

	return downloads, nil
}

func bundleResponses(t *torrent.Torrent, downloadPlan preview.DownloadPlan) ([]preview.DownloadedPart, error) {
	downloads := make([]preview.DownloadedPart, 0)
	for _, plan := range downloadPlan.GetPlan() {
		piece := new(bytes.Buffer)
		for pieceIdx := plan.Start(); pieceIdx <= plan.End(); pieceIdx++ {
			buf := make([]byte, plan.EndOffset(pieceIdx))
			off := plan.StartOffset(pieceIdx)
			_, err := t.Piece(pieceIdx).Storage().ReadAt(buf, int64(off))
			// TODO: All those checks fail with Rene torrent.
			/*if n != len(buf) {
				return nil, fmt.Errorf("unable to read all the piece block. expected %v, having %v", len(buf), n)
			}*/
			if err != nil {
				return nil, err
			}
			_, err = piece.Write(buf)
			if err != nil {
				return nil, err
			}
		}
		download := preview.NewDownloadedPart(t.InfoHash().HexString(), plan, piece.Bytes())
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
	uniquePartsWaitingFor := make(map[int]interface{})
	for _, plan := range downloadPlan.GetPlan() {
		for pIdx := plan.Start(); pIdx <= plan.End(); pIdx++ {
			if t.Piece(pIdx).State().Complete {
				continue
			}
			uniquePartsWaitingFor[pIdx] = struct{}{}
		}
	}
	return len(uniquePartsWaitingFor)
}

func downloadPieces(t *torrent.Torrent, downloadPlan preview.DownloadPlan) {
	// Idempotent. All the pieces already downloaded are ignored.
	for _, plan := range downloadPlan.GetPlan() {
		t.DownloadPieces(plan.Start(), plan.End()+1)
	}
}

func (r *TorrentClient) waitPiecesToDownload(ctx context.Context, t *torrent.Torrent, waitingFor int) {
	for waitingFor > 0 {
		select {
		case _v := <-t.SubscribePieceStateChanges().Values:
			v, ok := _v.(torrent.PieceStateChange)
			if !ok {
				continue
			}

			if v.Complete {
				waitingFor--
				r.logger.WithFields(
					logrus.Fields{"pieceIdx": v.Index,
						"complete":   v.Complete,
						"waitingFor": waitingFor,
						"torrent":    t.Name(),
					},
				).Info("piece download completed")
			}
		case <-time.After(time.Second * 3):
			r.logger.WithFields(
				logrus.Fields{
					"peersCount": len(t.PeerConns()),
					"torrent":    t.Name(),
					"piecesLeft": waitingFor,
				},
			).Debug("number of connected peers")
		case <-ctx.Done():
			break
		}
	}
}
