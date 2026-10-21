package service

import (
	"context"
	"time"

	"github.com/VsemoGosi-Adventure/events/internal/domain"
	"github.com/VsemoGosi-Adventure/events/internal/dto"
	"github.com/VsemoGosi-Adventure/events/internal/repo/db"
	"github.com/VsemoGosi-Adventure/events/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventService struct {
	storage *storage.Storage
}

func NewEventService(s *storage.Storage) *EventService {
	return &EventService{storage: s}
}

func (s *EventService) GetNearbyEvents(ctx context.Context, filter domain.GeoFilter) ([]dto.GeoEventItemResponse, error) {
	if filter.RadiusMeters <= 0 {
		filter.RadiusMeters = 3000
	}
	if filter.RadiusMeters < 100 || filter.RadiusMeters > 10000 {
		return nil, domain.ErrInvalidRadius
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	rows, err := s.storage.Queries.GetNearbyEvents(ctx, db.GetNearbyEventsParams{
		UserLng:      filter.UserLng,
		UserLat:      filter.UserLat,
		RadiusMeters: float64(filter.RadiusMeters),
		LimitVal:     int32(filter.Limit),
		OffsetVal:    int32(filter.Offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.GeoEventItemResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.GeoEventItemResponse{
			ID:              uuid.UUID(r.ID.Bytes),
			Title:           r.Title,
			Description:     pgTextToPtr(r.Description),
			StartsAt:        r.StartsAt.Time,
			EndsAt:          pgTimeToPtr(r.EndsAt),
			MaxParticipants: pgInt4ToPtr(r.MaxParticipants),
			DistanceMeters:  int(r.DistanceMeters),
			CurrentMembers:  int(r.CurrentMembers),
			LocationName:    r.LocName,
			Address:         r.LocAddress,
			Latitude:        r.LocLat,
			Longitude:       r.LocLng,
		})
	}

	return items, nil
}

func (s *EventService) GetEventByID(ctx context.Context, eventID uuid.UUID, userID *uuid.UUID) (*dto.EventDetailResponse, error) {
	row, err := s.storage.Queries.GetEventByID(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
	if err != nil {
		return nil, domain.ErrEventNotFound
	}

	isParticipant := false
	if userID != nil {
		status, err := s.storage.Queries.GetParticipantStatus(ctx, db.GetParticipantStatusParams{
			EventID: pgtype.UUID{Bytes: eventID, Valid: true},
			UserID:  pgtype.UUID{Bytes: *userID, Valid: true},
		})
		if err == nil && status == "joined" {
			isParticipant = true
		}
	}

	return &dto.EventDetailResponse{
		ID:              uuid.UUID(row.ID.Bytes),
		Title:           row.Title,
		Description:     pgTextToPtr(row.Description),
		StartsAt:        row.StartsAt.Time,
		EndsAt:          pgTimeToPtr(row.EndsAt),
		Status:          row.Status,
		MaxParticipants: pgInt4ToPtr(row.MaxParticipants),
		CurrentMembers:  int(row.CurrentMembers),
		IsParticipant:   isParticipant,
		LocationName:    row.LocName,
		Address:         row.LocAddress,
		Latitude:        row.LocLat,
		Longitude:       row.LocLng,
		ExternalURL:     pgTextToPtr(row.ExternalUrl),
	}, nil
}

func (s *EventService) JoinEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	row, err := s.storage.Queries.GetEventByID(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
	if err != nil {
		return domain.ErrEventNotFound
	}

	if row.Status != "active" {
		return domain.ErrEventCancelled
	}

	if row.MaxParticipants.Valid && int(row.CurrentMembers) >= int(row.MaxParticipants.Int32) {
		return domain.ErrEventFull
	}

	err = s.storage.Queries.UpsertParticipant(ctx, db.UpsertParticipantParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	})
	return err
}

func (s *EventService) LeaveEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	err := s.storage.Queries.LeaveParticipant(ctx, db.LeaveParticipantParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	})
	return err
}

func (s *EventService) GetMyEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.GeoEventItemResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.storage.Queries.GetMyEvents(ctx, db.GetMyEventsParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		LimitVal:  int32(limit),
		OffsetVal: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.GeoEventItemResponse, 0, len(rows))
	for _, r := range rows {
		lat := r.LocLat
		lng := r.LocLng

		items = append(items, dto.GeoEventItemResponse{
			ID:              uuid.UUID(r.ID.Bytes),
			Title:           r.Title,
			Description:     pgTextToPtr(r.Description),
			StartsAt:        r.StartsAt.Time,
			EndsAt:          pgTimeToPtr(r.EndsAt),
			MaxParticipants: pgInt4ToPtr(r.MaxParticipants),
			CurrentMembers:  int(r.CurrentMembers),
			LocationName:    r.LocName,
			Address:         r.LocAddress,
			Latitude:        lat,
			Longitude:       lng,
		})
	}

	return items, nil
}

func (s *EventService) CreatePing(ctx context.Context, eventID, authorID uuid.UUID, req dto.CreatePingRequest) (*dto.PingResponse, error) {
	status, err := s.storage.Queries.GetParticipantStatus(ctx, db.GetParticipantStatusParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: authorID, Valid: true},
	})
	if err != nil || status != "joined" {
		return nil, domain.ErrNotParticipant
	}

	if req.RadiusMeters < 100 || req.RadiusMeters > 10000 {
		return nil, domain.ErrInvalidRadius
	}

	// TODO возможность делать разное
	expiresAt := time.Now().Add(30 * time.Minute)

	p, err := s.storage.Queries.CreatePing(ctx, db.CreatePingParams{
		EventID:      pgtype.UUID{Bytes: eventID, Valid: true},
		AuthorID:     pgtype.UUID{Bytes: authorID, Valid: true},
		Message:      req.Message,
		RadiusMeters: int32(req.RadiusMeters),
		ExpiresAt:    pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// TODO: Добавить пуши как подключу панду

	return &dto.PingResponse{
		ID:           uuid.UUID(p.ID.Bytes),
		EventID:      uuid.UUID(p.EventID.Bytes),
		AuthorID:     uuid.UUID(p.AuthorID.Bytes),
		Message:      p.Message,
		RadiusMeters: int(p.RadiusMeters),
		ExpiresAt:    p.ExpiresAt.Time,
		CreatedAt:    p.CreatedAt.Time,
	}, nil
}

func (s *EventService) GetParticipants(ctx context.Context, eventID, requesterID uuid.UUID) ([]dto.ParticipantResponse, error) {
	status, err := s.storage.Queries.GetParticipantStatus(ctx, db.GetParticipantStatusParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: requesterID, Valid: true},
	})
	if err != nil || status != "joined" {
		return nil, domain.ErrNotParticipant
	}

	rows, err := s.storage.Queries.ListEventParticipants(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]dto.ParticipantResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, dto.ParticipantResponse{
			UserID:   uuid.UUID(r.UserID.Bytes),
			Status:   r.Status,
			JoinedAt: r.JoinedAt.Time,
		})
	}

	return result, nil
}

func (s *EventService) GetNearbyPings(ctx context.Context, lat, lng float64, limit, offset int) ([]dto.PingResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.storage.Queries.GetNearbyPings(ctx, db.GetNearbyPingsParams{
		UserLng:   lng,
		UserLat:   lat,
		LimitVal:  int32(limit),
		OffsetVal: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.PingResponse, 0, len(rows))
	for _, r := range rows {
		latVal := r.Lat
		lngVal := r.Lng

		result = append(result, dto.PingResponse{
			ID:             uuid.UUID(r.ID.Bytes),
			EventID:        uuid.UUID(r.EventID.Bytes),
			EventTitle:     r.EventTitle,
			AuthorID:       uuid.UUID(r.AuthorID.Bytes),
			Message:        r.Message,
			RadiusMeters:   int(r.RadiusMeters),
			DistanceMeters: int(r.DistanceMeters),
			Latitude:       latVal,
			Longitude:      lngVal,
			ExpiresAt:      r.ExpiresAt.Time,
			CreatedAt:      r.CreatedAt.Time,
		})
	}

	return result, nil
}

func (s *EventService) CheckAccess(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	status, err := s.storage.Queries.GetParticipantStatus(ctx, db.GetParticipantStatusParams{
		EventID: pgtype.UUID{Bytes: eventID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return false, nil
	}

	return status == "joined", nil
}

func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgTimeToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgInt4ToPtr(t pgtype.Int4) *int {
	if !t.Valid {
		return nil
	}
	v := int(t.Int32)
	return &v
}
