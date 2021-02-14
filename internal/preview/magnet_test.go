package preview_test

import (
	"github.com/stretchr/testify/assert"
	"prevtorrent/internal/preview"
	"testing"
)

func Test_Magnet_New(t *testing.T) {
	mag := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	m, err := preview.NewMagnet(mag)
	assert.NoError(t, err)
	assert.Equal(t, mag, m.Value())
	assert.Equal(t, "zocmzqipffw7ollmic5hub6bpcsdeoqu", m.ID())
}

func Test_Magnet_Invalid(t *testing.T) {
	mag := "invalid"
	_, err := preview.NewMagnet(mag)
	assert.Error(t, err)
}
