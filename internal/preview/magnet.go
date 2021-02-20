package preview

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var magnetValidationRegexp = regexp.MustCompile(`magnet:\?xt=urn:btih:([a-zA-Z0-9]*)`)

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=MagnetClient
type MagnetClient interface {
	Resolve(context.Context, Magnet) (Info, error)
}

// Magnet represents a magnet
type Magnet struct {
	id    string // the identifier of the magnet
	value string // the full URI like: magnet:?xt=....
}

var ErrInvalidMagnetFormat = errors.New("invalid magnet")

// NewMagnet returns a Magnet
func NewMagnet(value string) (Magnet, error) {
	if !magnetValidationRegexp.Match([]byte(value)) {
		return Magnet{}, ErrInvalidMagnetFormat
	}
	id := magnetValidationRegexp.FindStringSubmatch(value)[1]
	value = strings.ReplaceAll(value, id, strings.ToUpper(id))
	id = strings.ToLower(id)

	if !(len(id) == 32 || len(id) == 40) {
		return Magnet{}, errors.New("id must have 32 chars (hex encoded) or 40 chars (base32 encoded)")
	}
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
