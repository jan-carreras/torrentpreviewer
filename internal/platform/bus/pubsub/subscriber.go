package pubsub

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
	"prevtorrent/kit/command"
)

type Message struct {
	msg *message.Message
}

func NewMessage(msg *message.Message) *Message {
	return &Message{msg: msg}
}

func (m Message) UUID() command.UUID {
	return command.UUID(m.msg.UUID)
}

func (m Message) Payload() command.Payload {
	return command.Payload(m.msg.Payload)
}

func (m Message) ACK() bool {
	return m.msg.Ack()
}

type Subscriber struct {
	client message.Subscriber
	ch     chan command.Message
}

func NewSubscriber(client message.Subscriber) (*Subscriber, error) {
	return &Subscriber{client: client, ch: make(chan command.Message, 1)}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan command.Message, error) {
	messageCh, err := s.client.Subscribe(ctx, "command.downloadPartials")
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, isOpen := <-messageCh:
				if !isOpen {
					return
				}
				s.ch <- NewMessage(msg)
			}
		}
	}()

	return s.ch, nil

}

func (s *Subscriber) Close() error {
	closed := s.client.Close()
	close(s.ch)
	return closed
}
