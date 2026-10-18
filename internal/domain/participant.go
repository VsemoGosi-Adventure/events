package domain

import (
	"time"

	"github.com/google/uuid"
)

type ParticipantStatus string

const (
	ParticipantStatusJoined ParticipantStatus = "joined"
	ParticipantStatusLeft   ParticipantStatus = "left"
)

type Participant struct {
	ID        uuid.UUID         `json:"id"`
	EventID   uuid.UUID         `json:"event_id"`
	UserID    uuid.UUID         `json:"user_id"`
	Status    ParticipantStatus `json:"status"`
	JoinedAt  time.Time         `json:"joined_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
