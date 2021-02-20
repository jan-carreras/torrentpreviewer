package preview_test

import (
	"prevtorrent/internal/preview"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Magnet_New(t *testing.T) {
	mag := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	m, err := preview.NewMagnet(mag)
	assert.NoError(t, err)
	assert.Equal(t, mag, m.Value())
	assert.Equal(t, "cb84ccc10f296df72d6c40ba7a07c178a4323a14", m.ID())
}

func Test_Magnet_NewFixesCase(t *testing.T) {
	mag := "magnet:?xt=urn:btih:zOcmZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	m, err := preview.NewMagnet(mag)
	assert.NoError(t, err)
	assert.Equal(t, "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU", m.Value())
	assert.Equal(t, "cb84ccc10f296df72d6c40ba7a07c178a4323a14", m.ID())
}
func Test_Magnet_ConvertToHexEncoding(t *testing.T) {
	mag := "magnet:?xt=urn:btih:3htucd42zxr3iigmr2dq6hibmlmlzhex"

	m, err := preview.NewMagnet(mag)
	assert.NoError(t, err)
	assert.Equal(t, "magnet:?xt=urn:btih:3HTUCD42ZXR3IIGMR2DQ6HIBMLMLZHEX", m.Value())
	assert.Equal(t, "d9e7410f9acde3b420cc8e870f1d0162d8bc9c97", m.ID())
}

func Test_Magnet_Invalid(t *testing.T) {
	mag := "invalid"
	_, err := preview.NewMagnet(mag)
	assert.Error(t, err)

	mag = "magnet:?xt=urn:btih:1234567890"
	_, err = preview.NewMagnet(mag)
	assert.Error(t, err)
}
