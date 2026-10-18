package domain

import (
	"context"

	"github.com/google/uuid"
)

type GeoFilter struct {
	UserLat      float64
	UserLng      float64
	RadiusMeters int
	TagIDs       []int
	Limit        int
	Offset       int
}

type GeoItem struct {
	Event          Event
	DistanceMeters float64
	CurrentMembers int
}

type EventRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
	GetNearbyGeo(ctx context.Context, filter GeoFilter) ([]GeoItem, error)
	AddParticipant(ctx context.Context, eventID, userID uuid.UUID) error
	RemoteParticipant(ctx context.Context, eventID, userID uuid.UUID) error
	GetParticipantsCount(ctx context.Context, eventID uuid.UUID) (int, error)
	IsParticipant(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

type LocationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Location, error)
	UpsertByExternalID(ctx context.Context, loc *Location) error
}

type PingRepository interface {
	Create(ctx context.Context, ping *Ping) error
	GetActiveByEventID(ctx context.Context, eventID uuid.UUID) ([]Ping, error)
}
