package bittorrentproto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

func (r *TorrentClient) DownloadParts(ctx context.Context, downloadPlan preview.DownloadPlan) (chan preview.Piece, error) {
	t, err := r.getTorrent(downloadPlan)
	if err != nil {
		return nil, err
	}

	output := make(chan preview.Piece, downloadPlan.CountPieces())

	startTorrentDownload(t, downloadPlan)

	r.publishPartsThatWeAlreadyHave(t, output, downloadPlan)
	go r.waitPiecesToDownload(ctx, output, t, downloadPlan)

	return output, nil
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

func (r *TorrentClient) publishPartsThatWeAlreadyHave(t *torrent.Torrent, outputCh chan preview.Piece, downloadPlan preview.DownloadPlan) {
	for _, plan := range downloadPlan.GetPlan() {
		for pIdx := plan.Start(); pIdx <= plan.End(); pIdx++ {
			if t.Piece(pIdx).State().Complete {
				buf := r.readPiece(t, pIdx)
				outputCh <- preview.NewPiece(t.InfoHash().HexString(), pIdx, buf)
			}
		}
	}
}

func startTorrentDownload(t *torrent.Torrent, downloadPlan preview.DownloadPlan) {
	// Idempotent. All the pieces already downloaded are ignored.
	for _, plan := range downloadPlan.GetPlan() {
		t.DownloadPieces(plan.Start(), plan.End()+1) //  (start, end]
	}
}

func (r *TorrentClient) waitPiecesToDownload(ctx context.Context, outputCh chan preview.Piece, t *torrent.Torrent, downloadPlan preview.DownloadPlan) {
	// TODO: If we don't have any peer for a while we might disconnect as well

	defer close(outputCh)
	waitingFor := countNumberPiecesWaitingFor(t, downloadPlan)
	for waitingFor > 0 {
		select {
		case _v := <-t.SubscribePieceStateChanges().Values:
			v, ok := _v.(torrent.PieceStateChange)
			if !ok {
				continue
			}

			if v.Complete {
				waitingFor--
				buf := r.readPiece(t, v.Index)
				if buf == nil {
					continue
				}
				outputCh <- preview.NewPiece(t.InfoHash().HexString(), v.Index, buf)

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

func (r *TorrentClient) readPiece(t *torrent.Torrent, idx int) []byte {
	buf := make([]byte, t.Info().PieceLength)
	n, err := t.Piece(idx).Storage().ReadAt(buf, 0)
	if err != nil {
		r.logger.WithFields(
			logrus.Fields{"pieceIdx": idx,
				"complete": t.Piece(idx).Storage().Completion().Complete,
				"torrent":  t.Name(),
				"error":    err,
			},
		).Error("unknown error when reading the piece from storage")
		return nil
	}

	if n != len(buf) {
		r.logger.WithFields(
			logrus.Fields{"pieceIdx": idx,
				"complete":        t.Piece(idx).Storage().Completion().Complete,
				"torrent":         t.Name(),
				"expectingToRead": len(buf),
				"havingRead":      n,
			},
		).Error("unable to read piece from storage. bytes mismatch")
		return nil
	}
	return buf
}

func bundleResponses(t *torrent.Torrent, downloadPlan preview.DownloadPlan) ([]preview.MediaPart, error) {
	downloads := make([]preview.MediaPart, 0)
	for _, plan := range downloadPlan.GetPlan() {
		piece := new(bytes.Buffer)
		for pieceIdx := plan.Start(); pieceIdx <= plan.End(); pieceIdx++ {
			buf := make([]byte, plan.EndOffset(pieceIdx))
			off := plan.StartOffset(pieceIdx)
			n, err := t.Piece(pieceIdx).Storage().ReadAt(buf, int64(off))
			if n != len(buf) {
				return nil, fmt.Errorf("unable to read all the piece block. expected %v, having %v", len(buf), n)
			}
			if err != nil {
				return nil, err
			}
			_, err = piece.Write(buf)
			if err != nil {
				return nil, err
			}
		}
		download := preview.NewMediaPart(t.InfoHash().HexString(), plan, piece.Bytes())
		downloads = append(downloads, download)
	}
	return downloads, nil
}
