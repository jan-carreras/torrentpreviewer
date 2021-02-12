package preview

import (
	"context"
	"errors"
	"regexp"
)

var magnetValidationRegexp = regexp.MustCompile("magnet:\\?xt=urn:btih:[a-zA-Z0-9]*")

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=MagnetClient
type MagnetClient interface {
	Resolve(context.Context, Magnet) ([]byte, error)
	DownloadParts(context.Context, DownloadPlan) ([]DownloadedPart, error)
}

type Magnet struct {
	value string
}

var ErrInvalidMagnetFormat = errors.New("invalid magnet")

func NewMagnet(value string) (Magnet, error) {
	if !magnetValidationRegexp.Match([]byte(value)) {
		return Magnet{}, ErrInvalidMagnetFormat
	}
	return Magnet{value: value}, nil
}

func (m Magnet) Value() string {
	return m.value
}