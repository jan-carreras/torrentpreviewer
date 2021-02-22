package bittorrentproto

import (
	"bytes"
	"context"
	"errors"
	"prevtorrent/internal/preview"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sirupsen/logrus"
)

const (
	maxDownloadTime = time.Minute * 15
	seederWaitTime  = time.Second * 30
)

type TorrentClient struct {
	client *torrent.Client
	logger *logrus.Logger
}

func NewTorrentClient(client *torrent.Client, logger *logrus.Logger) *TorrentClient {
	return &TorrentClient{client: client, logger: logger}
}

func (r *TorrentClient) Resolve(ctx context.Context, m preview.Magnet) (preview.Info, error) {
	t, err := r.client.AddMagnet(m.Value())
	if err != nil {
		return preview.Info{}, err
	}

	if err := r.waitForInfo(ctx, t); err != nil {
		return preview.Info{}, err
	}
	return parseTorrent(t)
}

func parseTorrent(t *torrent.Torrent) (preview.Info, error) {
	buf := new(bytes.Buffer)
	err := t.Metainfo().Write(buf)
	if err != nil {
		return preview.Info{}, err
	}

	files := make([]preview.FileInfo, 0)
	for idx, f := range t.Info().Files {
		fi, err := preview.NewFileInfo(idx, int(f.Length), f.DisplayPath(t.Info()))
		if err != nil {
			return preview.Info{}, err
		}

		files = append(files, fi)
	}

	return preview.NewInfo(
		t.Metainfo().HashInfoBytes().String(),
		t.Name(),
		int(t.Info().PieceLength),
		files,
		buf.Bytes(),
	)
}

func (r *TorrentClient) waitForInfo(ctx context.Context, t *torrent.Torrent) error {
	select {
	case <-t.GotInfo():
		return nil
	case <-ctx.Done():
		return errors.New("context cancelled while trying to get info")
	}
}

func (r *TorrentClient) DownloadParts(ctx context.Context, downloadPlan preview.DownloadPlan) (*preview.PieceRegistry, error) {
	storage := preview.NewPieceInMemoryStorage(downloadPlan)
	registry, err := preview.NewPieceRegistry(ctx, r.logger, &downloadPlan, storage)
	if err != nil {
		return nil, err
	}

	t, err := r.getTorrent(downloadPlan)
	if err != nil {
		return nil, err
	}

	startTorrentDownload(t, downloadPlan)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		wg.Wait()
		registry.NoMorePieces()
	}()
	go r.publishPartsThatWeAlreadyHave(wg, t, registry, downloadPlan)
	go r.waitPiecesToDownload(ctx, wg, registry, t, downloadPlan)

	return registry, nil
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

func (r *TorrentClient) publishPartsThatWeAlreadyHave(wg *sync.WaitGroup, t *torrent.Torrent, registry *preview.PieceRegistry, downloadPlan preview.DownloadPlan) {
	defer wg.Done()
	for _, plan := range downloadPlan.GetPlan() {
		for pIdx := plan.Start(); pIdx <= plan.End(); pIdx++ {
			if t.Piece(pIdx).State().Complete {
				buf := r.readPiece(t, pIdx)
				registry.RegisterPiece(preview.NewPiece(t.InfoHash().HexString(), pIdx, buf))
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

func (r *TorrentClient) waitPiecesToDownload(ctx context.Context, wg *sync.WaitGroup, registry *preview.PieceRegistry, t *torrent.Torrent, downloadPlan preview.DownloadPlan) {
	defer wg.Done()
	defer t.Drop() // Delete all the chunks we have in the storage

	waitingFor := countNumberPiecesWaitingFor(t, downloadPlan)
	if waitingFor == 0 {
		r.logger.WithFields(
			logrus.Fields{
				"waitingFor": 0,
				"torrent":    t.Name(),
			},
		).Debug("all pieces already downloaded")
	}

	if !r.hasSeeders(ctx, t, seederWaitTime) {
		return
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, maxDownloadTime)
	defer cancel()

	subscription := t.SubscribePieceStateChanges()
	defer subscription.Close()
	for waitingFor > 0 {
		select {
		case _v, isOpen := <-subscription.Values:
			if !isOpen {
				r.logger.WithFields(
					logrus.Fields{
						"waitingFor": waitingFor,
						"torrent":    t.Name(),
					},
				).Info("transmissions subscriber closed")
			}
			v, ok := _v.(torrent.PieceStateChange)
			if !ok || !v.Complete {
				continue
			}

			waitingFor--
			buf := r.readPiece(t, v.Index)
			if buf == nil {
				continue
			}
			registry.RegisterPiece(preview.NewPiece(t.InfoHash().HexString(), v.Index, buf))

			r.logger.WithFields(
				logrus.Fields{"pieceIdx": v.Index,
					"complete":   v.Complete,
					"waitingFor": waitingFor,
					"torrent":    t.Name(),
				},
			).Info("piece download completed")

		case <-time.After(time.Second * 3):
			r.logger.WithFields(
				logrus.Fields{
					"seedersCount":     t.Stats().ConnectedSeeders,
					"piecesLeft":       waitingFor,
					"activePeers":      t.Stats().ActivePeers,
					"chunksReadUseful": t.Stats().ChunksReadUseful,
					"ChunksReadWasted": t.Stats().ChunksReadWasted,
					"torrent":          t.Name(),
				},
			).Debug("number of connected peers")
		case <-ctxTimeout.Done():
			r.logger.WithFields(
				logrus.Fields{
					"peersCount": len(t.PeerConns()),
					"torrent":    t.Name(),
					"piecesLeft": waitingFor,
					"context":    ctxTimeout.Err(),
				},
			).Error("goroutine stopped because context closed")
			return
		}
	}
}

func (r *TorrentClient) hasSeeders(ctx context.Context, t *torrent.Torrent, duration time.Duration) bool {
	ctxSeederTimeout, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	for t.Stats().ConnectedSeeders == 0 {
		select {
		case <-ctxSeederTimeout.Done():
			r.logger.WithFields(
				logrus.Fields{
					"torrent":     t.InfoHash().String(),
					"torrentName": t.Name(),
					"context":     ctxSeederTimeout.Err(),
				},
			).Error("torrent without seeders")
			return false
		case <-time.After(time.Second):
		}
	}
	return true
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

func (r *TorrentClient) Import(ctx context.Context, raw []byte) (preview.Info, error) {
	data := bytes.NewBuffer(raw)
	metaInfo, err := metainfo.Load(data)
	if err != nil {
		return preview.Info{}, err
	}
	t, err := r.client.AddTorrent(metaInfo)
	if err != nil {
		return preview.Info{}, err
	}

	return parseTorrent(t)
}
