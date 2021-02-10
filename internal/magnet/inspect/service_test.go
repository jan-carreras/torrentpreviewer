package inspect_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"prevtorrent/internal/magnet"
	"prevtorrent/internal/magnet/inspect"
	"prevtorrent/internal/magnet/platform/client/clientmocks"
	"testing"
)

func Test_MagnetService_Inspect_Succeed(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	mag, err := magnet.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	repo := new(clientmocks.FileDescriber)
	repo.On("DescribeFiles", mock.Anything, mag).Return(nil, nil)

	s := inspect.NewService(repo)
	err = s.Inspect(context.Background(), inspect.ServiceCMD{Magnet: inputMagnet})
	assert.NoError(t, err)
}

func Test_MagnetService_Inspect_RepositoryError(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	mag, err := magnet.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	repo := new(clientmocks.FileDescriber)
	repo.On("DescribeFiles", mock.Anything, mag).Return(nil, errors.New("fake repo error"))

	s := inspect.NewService(repo)
	err = s.Inspect(context.Background(), inspect.ServiceCMD{Magnet: inputMagnet})
	assert.Error(t, err)
}

func Test_MagnetService_Inspect_InvalidMagnetError(t *testing.T) {
	inputMagnet := "invalid magnet"

	repo := new(clientmocks.FileDescriber)

	s := inspect.NewService(repo)
	err := s.Inspect(context.Background(), inspect.ServiceCMD{Magnet: inputMagnet})
	assert.Error(t, err)
}


