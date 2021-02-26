package bus

import (
	"context"
)

//go:generate mockery --case=snake --outpkg=busmocks --output=busmocks --name=Event
type Event interface {
	Publish(ctx context.Context, event interface{}) error
}

//go:generate mockery --case=snake --outpkg=busmocks --output=busmocks --name=Command
type Command interface {
	Send(ctx context.Context, cmd interface{}) error
}
