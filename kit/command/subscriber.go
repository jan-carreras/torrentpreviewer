package command

import "context"

type UUID string
type Payload []byte

type Message interface {
	UUID() UUID
	Payload() Payload
	ACK() bool
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan Message, error)
	Close() error
}
