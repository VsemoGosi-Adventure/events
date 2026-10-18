package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusActive    EventStatus = "active"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusFinished  EventStatus = "finished"
)

type Event struct {
	ID              uuid.UUID   `json:"id"`
	LocationID      uuid.UUID   `json:"location_id"`
	Location        *Location   `json:"location,omitempty"`
	Title           string      `json:"title"`
	Description     *string     `json:"description,omitempty"`
	StartsAt        time.Time   `json:"starts_at"`
	EndsAt          *time.Time  `json:"ends_at,omitempty"`
	MaxParticipants *int        `json:"max_participants,omitempty"`
	Status          EventStatus `json:"status"`
	ExternalURL     *string     `json:"external_url,omitempty"`
	TagIDs          []int       `json:"tag_ids,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (e *Event) CanJoin(currentParticipantsCount int) error {
	if e.Status == EventStatusCancelled {
		return ErrEventCancelled
	}
	if e.Status == EventStatusFinished || (e.EndsAt != nil && e.EndsAt.Before(time.Now())) {
		return ErrEventFinished
	}
	if e.MaxParticipants != nil && currentParticipantsCount >= *e.MaxParticipants {
		return ErrEventFull
	}
	return nil
}
