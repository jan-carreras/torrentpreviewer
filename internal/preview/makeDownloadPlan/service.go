package makeDownloadPlan

import (
	"context"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"

	"github.com/sirupsen/logrus"
)

const (
	mb = 1 << (10 * 2) // MiB, really
)

type Service struct {
	logger            *logrus.Logger
	commandBus        bus.Command
	torrentRepository preview.TorrentRepository
	imageRepository   preview.ImageRepository
}

func NewService(
	logger *logrus.Logger,
	commandBus bus.Command,
	torrentRepository preview.TorrentRepository,
	imageRepository preview.ImageRepository,
) Service {
	return Service{
		logger:            logger,
		commandBus:        commandBus,
		torrentRepository: torrentRepository,
		imageRepository:   imageRepository,
	}
}

func (s Service) Download(ctx context.Context, cmd CMD) error {
	torrent, err := s.torrentRepository.Get(ctx, cmd.TorrentID)
	if err != nil {
		return err
	}

	plan, err := s.makePlan(ctx, torrent)
	if err != nil {
		return err
	}

	downloadCMD, err := s.makeDownloadPartialCommands(plan)
	if err != nil {
		return err
	}

	if err := s.executeCMDs(ctx, downloadCMD); err != nil {
		return err
	}

	return nil
}

func (s Service) makePlan(ctx context.Context, t preview.Torrent) (*preview.DownloadPlan, error) {
	torrentImages, err := s.imageRepository.ByTorrent(ctx, t.ID())
	if err != nil {
		return nil, err
	}

	plan := preview.NewDownloadPlan(t)
	if err := plan.AddAll(torrentImages); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s Service) makeDownloadPartialCommands(plan *preview.DownloadPlan) ([]downloadPartials.CMD, error) {
	plans, err := plan.GetCappedPlans(downloadSize(plan.GetTorrent()))
	if err != nil {
		return nil, err
	}

	commands := make([]downloadPartials.CMD, 0)
	for _, partialPlan := range plans {
		files := make([]downloadPartials.File, 0, len(partialPlan))
		for _, fileRange := range partialPlan {
			files = append(files, downloadPartials.File{
				FileID: fileRange.FileID(),
				Start:  fileRange.FileStart(),
				Length: fileRange.FileLength(),
			})
		}

		commands = append(commands, downloadPartials.CMD{ID: plan.GetTorrent().ID(), Files: files})
	}
	return commands, nil
}

func (s Service) executeCMDs(ctx context.Context, commands []downloadPartials.CMD) error {
	for _, downloadCMD := range commands {
		if err := s.commandBus.Send(ctx, downloadCMD); err != nil {
			return err
		}
	}

	return nil
}

func downloadSize(t preview.Torrent) int {
	downloadSizePerPlan := 100 * mb
	pieces := downloadSizePerPlan / t.PieceLength()
	return t.PieceLength() * pieces
}
