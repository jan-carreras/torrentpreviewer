package inspect

import (
	"context"
	"prevtorrent/internal/magnet"
)

type Service struct {
	fileDescriberRepository magnet.FileDescriber
}

func NewService(fd magnet.FileDescriber) *Service {
	return &Service{fileDescriberRepository: fd}
}

type ServiceCMD struct {
	Magnet string
}

func (s *Service) Inspect(ctx context.Context, cmd ServiceCMD) error {
	m, err := magnet.NewMagnet(cmd.Magnet)
	if err != nil {
		return err
	}

	if _, err := s.fileDescriberRepository.GetMagnetInfo(ctx, m); err != nil {
		return err
	}
	return nil
}
