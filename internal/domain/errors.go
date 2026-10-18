package domain

import "errors"

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventCancelled     = errors.New("event is cancelled")
	ErrEventFinished      = errors.New("event is already finished")
	ErrEventFull          = errors.New("event has reached maximum participants")
	ErrAlreadyParticipant = errors.New("user is already a participant")
	ErrNotParticipant     = errors.New("user is not a participant")
	ErrInvalidRadius      = errors.New("radius must be between 100 and 10000 meters")
	ErrPingExpired        = errors.New("ping has expired")
)
