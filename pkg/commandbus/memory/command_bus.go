package memory

import (
	"context"
	"fmt"

	messagebus "github.com/vardius/message-bus"

	"github.com/vardius/go-api-boilerplate/pkg/application"
	"github.com/vardius/go-api-boilerplate/pkg/commandbus"
	"github.com/vardius/go-api-boilerplate/pkg/domain"
	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/log"
)

// New creates in memory command bus
func New(maxConcurrentCalls int, logger *log.Logger) commandbus.CommandBus {
	return &commandBus{messagebus.New(maxConcurrentCalls), logger}
}

type commandBus struct {
	messageBus messagebus.MessageBus
	logger     *log.Logger
}

func (bus *commandBus) Publish(ctx context.Context, command domain.Command) error {
	out := make(chan error, 1)
	defer close(out)

	bus.logger.Debug(ctx, "[CommandBus] Publish: %s %+v\n", command.GetName(), command)
	bus.messageBus.Publish(command.GetName(), ctx, command, out)

	ctxDoneCh := ctx.Done()
	select {
	case <-ctxDoneCh:
		return errors.Wrap(fmt.Errorf("%w: %s", application.ErrTimeout, ctx.Err()))
	case err := <-out:
		if err != nil {
			return errors.Wrap(fmt.Errorf("create client failed: %w", err))
		}
		return nil
	}
}

func (bus *commandBus) Subscribe(ctx context.Context, commandName string, fn commandbus.CommandHandler) error {
	bus.logger.Info(nil, "[CommandBus] Subscribe: %s\n", commandName)

	// unsubscribe all other handlers
	bus.messageBus.Close(commandName)

	return bus.messageBus.Subscribe(commandName, func(ctx context.Context, command domain.Command, out chan<- error) {
		out <- fn(ctx, command)
	})
}

func (bus *commandBus) Unsubscribe(ctx context.Context, commandName string) error {
	bus.logger.Info(nil, "[CommandBus] Unsubscribe: %s\n", commandName)
	bus.messageBus.Close(commandName)

	return nil
}
