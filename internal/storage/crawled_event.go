package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/VsemoGosi-Adventure/events/internal/consumer"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpsertCrawledEvent saves a crawled event and its location atomically.
func (s *Storage) UpsertCrawledEvent(ctx context.Context, event consumer.CrawledEventMessage) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin crawled event transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// The payload has no location ID. A stable event-based key lets re-deliveries
	// update the same location without creating orphan rows.
	key := sha256.Sum256([]byte(event.ExternalSource + "\x00" + event.ExternalSourceID))
	locationKey := "crawled:" + hex.EncodeToString(key[:])

	var locationID pgtype.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO events.locations (name, address, coordinates, external_source_id)
VALUES ($1, $2, ST_SetSRID(ST_MakePoint($4::float8, $3::float8), 4326)::geography, $5)
ON CONFLICT (external_source_id) DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    coordinates = EXCLUDED.coordinates,
    updated_at = NOW()
RETURNING id`, event.Location.Name, event.Location.Address,
		event.Location.Latitude, event.Location.Longitude, locationKey).Scan(&locationID)
	if err != nil {
		return fmt.Errorf("upsert crawled event location: %w", err)
	}

	tags := event.Tags
	if tags == nil {
		tags = []string{}
	}
	_, err = tx.Exec(ctx, `
INSERT INTO events.events
    (location_id, title, description, starts_at, ends_at, max_participants,
     external_url, external_source, external_source_id, join_policy, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (external_source, external_source_id) DO UPDATE SET
    location_id = EXCLUDED.location_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    starts_at = EXCLUDED.starts_at,
    ends_at = EXCLUDED.ends_at,
    max_participants = EXCLUDED.max_participants,
    external_url = EXCLUDED.external_url,
    join_policy = EXCLUDED.join_policy,
    tags = EXCLUDED.tags,
    updated_at = NOW()`,
		locationID, event.Title, nullableText(event.Description), event.StartsAt,
		event.EndsAt, event.MaxParticipants, nullableText(event.ExternalURL),
		event.ExternalSource, event.ExternalSourceID, event.JoinPolicy, tags)
	if err != nil {
		return fmt.Errorf("upsert crawled event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit crawled event: %w", err)
	}
	return nil
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
