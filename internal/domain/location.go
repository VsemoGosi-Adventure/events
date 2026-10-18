package domain

import (
	"time"

	"github.com/google/uuid"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Location struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Address          string      `json:"address"`
	Coordinates      Coordinates `json:"coordinates"`
	ExternalSourceID *string     `json:"external_source_id,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
