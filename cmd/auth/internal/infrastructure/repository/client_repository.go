/*
Package repository holds event sourced repositories
*/
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/vardius/go-api-boilerplate/cmd/auth/internal/domain/client"
	"github.com/vardius/go-api-boilerplate/pkg/application"
	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/eventbus"
	"github.com/vardius/go-api-boilerplate/pkg/eventstore"
)

type clientRepository struct {
	eventStore eventstore.EventStore
	eventBus   eventbus.EventBus
}

// Save current client changes to event store and publish each event with an event bus
func (r *clientRepository) Save(ctx context.Context, u client.Client) error {
	if err := r.eventStore.Store(ctx, u.Changes()); err != nil {
		return errors.Wrap(err)
	}

	for _, event := range u.Changes() {
		if err := r.eventBus.Publish(ctx, event); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

// Save current client changes to event store and publish each event with an event bus
// blocks until event handlers are finished
func (r *clientRepository) SaveAndAcknowledge(ctx context.Context, u client.Client) error {
	if err := r.eventStore.Store(ctx, u.Changes()); err != nil {
		return errors.Wrap(err)
	}

	for _, event := range u.Changes() {
		if err := r.eventBus.PublishAndAcknowledge(ctx, event); err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

// Get client with current state applied
func (r *clientRepository) Get(ctx context.Context, id uuid.UUID) (client.Client, error) {
	events, err := r.eventStore.GetStream(ctx, id, client.StreamName)
	if err != nil {
		return client.Client{}, errors.Wrap(err)
	}

	if len(events) == 0 {
		return client.Client{}, application.ErrNotFound
	}

	return client.FromHistory(events), nil
}

// NewClientRepository creates new client event sourced repository
func NewClientRepository(store eventstore.EventStore, bus eventbus.EventBus) client.Repository {
	return &clientRepository{store, bus}
}
