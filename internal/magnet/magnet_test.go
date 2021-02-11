package magnet_test

import (
	"github.com/stretchr/testify/assert"
	"prevtorrent/internal/magnet"
	"testing"
)

func Test_Magnet_New(t *testing.T) {
	mag := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	m, err := magnet.NewMagnet(mag)
	assert.NoError(t, err)
	assert.Equal(t, mag, m.Value())
}

func Test_Magnet_Invalid(t *testing.T) {
	mag := "invalid"
	_, err := magnet.NewMagnet(mag)
	assert.Error(t, err)
}
