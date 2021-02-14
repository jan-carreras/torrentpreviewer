package preview

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var magnetValidationRegexp = regexp.MustCompile("magnet:\\?xt=urn:btih:([a-zA-Z0-9]*)")

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=MagnetClient
type MagnetClient interface {
	Resolve(context.Context, Magnet) (Info, error)
	DownloadParts(context.Context, DownloadPlan) (*PieceRegistry, error) // TODO: It has nothing to do with Magnet, really
}

type Magnet struct {
	id    string
	value string
}

var ErrInvalidMagnetFormat = errors.New("invalid magnet")

func NewMagnet(value string) (Magnet, error) {
	if !magnetValidationRegexp.Match([]byte(value)) {
		return Magnet{}, ErrInvalidMagnetFormat
	}
	id := strings.ToLower(magnetValidationRegexp.FindStringSubmatch(value)[1])
	return Magnet{id: id, value: value}, nil
}

func (m Magnet) Value() string {
	return m.value
}

func (m Magnet) ID() string {
	return m.id
}
