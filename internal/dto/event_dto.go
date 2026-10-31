package dto

import (
	"time"

	"github.com/google/uuid"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// GeoEventItemResponse — элемент гео-выдачи и списка «Мои встречи»
type GeoEventItemResponse struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	StartsAt        time.Time  `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	MaxParticipants *int       `json:"max_participants,omitempty"`
	DistanceMeters  int        `json:"distance_meters,omitempty"`
	CurrentMembers  int        `json:"current_members"`
	LocationName    string     `json:"location_name"`
	Address         string     `json:"address"`
	Latitude        *float64   `json:"latitude,omitempty"`
	Longitude       *float64   `json:"longitude,omitempty"`
	Tags            []int      `json:"tags,omitempty"`
}

// EventDetailResponse — детальная карточка встречи
type EventDetailResponse struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	StartsAt        time.Time  `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	Status          string     `json:"status"`
	MaxParticipants *int       `json:"max_participants,omitempty"`
	CurrentMembers  int        `json:"current_members"`
	IsParticipant   bool       `json:"is_participant"`
	LocationName    string     `json:"location_name"`
	Address         string     `json:"address"`
	Latitude        *float64   `json:"latitude,omitempty"`
	Longitude       *float64   `json:"longitude,omitempty"`
	ExternalURL     *string    `json:"external_url,omitempty"`
	Tags            []int      `json:"tags"`
}

// ParticipantResponse — элемент списка участников встречи
type ParticipantResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Status   string    `json:"status"`
	JoinedAt time.Time `json:"joined_at"`
}

// CreatePingRequest — тело запроса для отправки «Пинга рядом»
type CreatePingRequest struct {
	Message      string `json:"message" binding:"required,max=255"`
	RadiusMeters int    `json:"radius_meters" binding:"required,min=100,max=10000"`
	TTL          int    `json:"ttl"`
}

// PingResponse — ответ с данными активного пинга
type PingResponse struct {
	ID             uuid.UUID `json:"id"`
	EventID        uuid.UUID `json:"event_id"`
	EventTitle     string    `json:"event_title,omitempty"`
	AuthorID       uuid.UUID `json:"author_id"`
	Message        string    `json:"message"`
	RadiusMeters   int       `json:"radius_meters"`
	DistanceMeters int       `json:"distance_meters,omitempty"`
	Latitude       float64   `json:"latitude,omitempty"`
	Longitude      float64   `json:"longitude,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// AccessCheckResponse —  проверка доступа к чату
type AccessCheckResponse struct {
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	HasAccess bool      `json:"has_access"`
}
