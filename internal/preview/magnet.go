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
}

// Magnet represents a magnet
type Magnet struct {
	id    string // the identifier of the magnet
	value string // the full URI like: magnet:\xt=....
}

var ErrInvalidMagnetFormat = errors.New("invalid magnet")

// NewMagnet returns a Magnet
func NewMagnet(value string) (Magnet, error) {
	if !magnetValidationRegexp.Match([]byte(value)) {
		return Magnet{}, ErrInvalidMagnetFormat
	}
	id := strings.ToLower(magnetValidationRegexp.FindStringSubmatch(value)[1])
	return Magnet{id: id, value: value}, nil
}

// Value returns the URI of the magnet
func (m Magnet) Value() string {
	return m.value
}

// ID returns the identified parsed from the URI of the magnet
func (m Magnet) ID() string {
	return m.id
}
