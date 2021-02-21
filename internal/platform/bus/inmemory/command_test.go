package inmemory_test

import (
	context "context"
	"errors"
	"io/ioutil"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/kit/command"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCommandBus(t *testing.T) {
	commandBus := inmemory.NewSyncCommandBus(fakeLogger())

	cmd := testCommand{}
	handler := fakeCommandHandler{}
	commandBus.Register(cmd.Type(), handler)

	err := commandBus.Dispatch(context.Background(), cmd)
	assert.Equal(t, fakeErr, err)
}

func TestCommanBus_UnknownCommand(t *testing.T) {
	commandBus := inmemory.NewSyncCommandBus(fakeLogger())

	cmd := testCommand{}
	err := commandBus.Dispatch(context.Background(), cmd)
	assert.NoError(t, err)
}

type testCommand struct{}

func (cmd testCommand) Type() command.Type {
	return "test-command"
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}

type fakeCommandHandler struct{}

var fakeErr = errors.New("fake error from handler")

func (f fakeCommandHandler) Handle(ctx context.Context, c command.Command) error {
	return fakeErr
}
