package domain

import (
	"time"

	"github.com/google/uuid"
)

type Ping struct {
	ID           uuid.UUID `json:"id"`
	EventID      uuid.UUID `json:"event_id"`
	AuthorID     uuid.UUID `json:"author_id"`
	Message      string    `json:"message"`
	RadiusMeters int       `json:"radius_meters"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p *Ping) IsActive() bool {
	return time.Now().Before(p.ExpiresAt)
}
