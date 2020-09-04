/*
Package user holds user domain logic
*/
package user

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/vardius/go-api-boilerplate/pkg/domain"
	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/identity"
)

// StreamName for user domain
var StreamName = fmt.Sprintf("%T", User{})

// User aggregate root
type User struct {
	id      uuid.UUID
	version int
	changes []domain.Event

	email EmailAddress
}

// New creates an User
func New() User {
	return User{}
}

// FromHistory loads current aggregate root state by applying all events in order
func FromHistory(events []domain.Event) User {
	u := New()

	for _, domainEvent := range events {
		var e domain.RawEvent

		switch domainEvent.Metadata.Type {
		case (AccessTokenWasRequested{}).GetType():
			accessTokenWasRequested := AccessTokenWasRequested{}
			if err := unmarshalPayload(domainEvent.Payload, &accessTokenWasRequested); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = accessTokenWasRequested
		case (EmailAddressWasChanged{}).GetType():
			emailAddressWasChanged := EmailAddressWasChanged{}
			if err := unmarshalPayload(domainEvent.Payload, &emailAddressWasChanged); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = emailAddressWasChanged
		case (WasRegisteredWithEmail{}).GetType():
			wasRegisteredWithEmail := WasRegisteredWithEmail{}
			if err := unmarshalPayload(domainEvent.Payload, &wasRegisteredWithEmail); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = wasRegisteredWithEmail
		case (WasRegisteredWithFacebook{}).GetType():
			wasRegisteredWithFacebook := WasRegisteredWithFacebook{}
			if err := unmarshalPayload(domainEvent.Payload, &wasRegisteredWithFacebook); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = wasRegisteredWithFacebook
		case (ConnectedWithFacebook{}).GetType():
			connectedWithFacebook := ConnectedWithFacebook{}
			if err := unmarshalPayload(domainEvent.Payload, &connectedWithFacebook); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = connectedWithFacebook
		case (WasRegisteredWithGoogle{}).GetType():
			wasRegisteredWithGoogle := WasRegisteredWithGoogle{}
			if err := unmarshalPayload(domainEvent.Payload, &wasRegisteredWithGoogle); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = wasRegisteredWithGoogle
		case (ConnectedWithGoogle{}).GetType():
			connectedWithGoogle := ConnectedWithGoogle{}
			if err := unmarshalPayload(domainEvent.Payload, &connectedWithGoogle); err != nil {
				log.Panicf("Error while trying to unmarshal user event %s. %s\n", domainEvent.Metadata.Type, err)
			}

			e = connectedWithGoogle
		default:
			log.Panicf("Unhandled user event %s\n", domainEvent.Metadata.Type)
		}

		u.transition(e)
		u.version++
	}

	return u
}

// ID returns aggregate root id
func (u User) ID() uuid.UUID {
	return u.id
}

// Version returns current aggregate root version
func (u User) Version() int {
	return u.version
}

// Changes returns all new applied events
func (u User) Changes() []domain.Event {
	return u.changes
}

// RegisterWithEmail alters current user state and append changes to aggregate root
func (u *User) RegisterWithEmail(ctx context.Context, id uuid.UUID, email EmailAddress) error {
	if _, err := u.trackChange(ctx, WasRegisteredWithEmail{
		ID:    id,
		Email: email,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// RegisterWithGoogle alters current user state and append changes to aggregate root
func (u *User) RegisterWithGoogle(ctx context.Context, id uuid.UUID, email EmailAddress, googleID, accessToken string) error {
	if _, err := u.trackChange(ctx, WasRegisteredWithGoogle{
		ID:          id,
		Email:       email,
		GoogleID:    googleID,
		AccessToken: accessToken,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// ConnectWithGoogle alters current user state and append changes to aggregate root
func (u *User) ConnectWithGoogle(ctx context.Context, googleID, accessToken string) error {
	if _, err := u.trackChange(ctx, ConnectedWithGoogle{
		ID:          u.id,
		GoogleID:    googleID,
		AccessToken: accessToken,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// RegisterWithFacebook alters current user state and append changes to aggregate root
func (u *User) RegisterWithFacebook(ctx context.Context, id uuid.UUID, email EmailAddress, facebookID, accessToken string) error {
	if _, err := u.trackChange(ctx, WasRegisteredWithFacebook{
		ID:          id,
		Email:       email,
		FacebookID:  facebookID,
		AccessToken: accessToken,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// ConnectWithFacebook alters current user state and append changes to aggregate root
func (u *User) ConnectWithFacebook(ctx context.Context, facebookID, accessToken string) error {
	if _, err := u.trackChange(ctx, ConnectedWithFacebook{
		ID:          u.id,
		FacebookID:  facebookID,
		AccessToken: accessToken,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// ChangeEmailAddress alters current user state and append changes to aggregate root
func (u *User) ChangeEmailAddress(ctx context.Context, email EmailAddress) error {
	if _, err := u.trackChange(ctx, EmailAddressWasChanged{
		ID:    u.id,
		Email: email,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// RequestAccessToken dispatches AccessTokenWasRequested event
func (u *User) RequestAccessToken(ctx context.Context) error {
	if _, err := u.trackChange(ctx, AccessTokenWasRequested{
		ID:    u.id,
		Email: u.email,
	}); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func (u *User) trackChange(ctx context.Context, e domain.RawEvent) (domain.Event, error) {
	u.transition(e)

	var (
		event domain.Event
		err   error
	)
	if i, hasIdentity := identity.FromContext(ctx); hasIdentity {
		event, err = domain.NewEvent(u.id, StreamName, u.version, e, &i)
	} else {
		event, err = domain.NewEvent(u.id, StreamName, u.version, e, nil)
	}
	if err != nil {
		return event, errors.Wrap(err)
	}

	u.changes = append(u.changes, event)

	return event, nil
}

func (u *User) transition(e domain.RawEvent) {
	switch e := e.(type) {
	case WasRegisteredWithEmail:
		u.id = e.ID
		u.email = e.Email
	case WasRegisteredWithGoogle:
		u.id = e.ID
		u.email = e.Email
	case WasRegisteredWithFacebook:
		u.id = e.ID
		u.email = e.Email
	case EmailAddressWasChanged:
		u.email = e.Email
	}
}
