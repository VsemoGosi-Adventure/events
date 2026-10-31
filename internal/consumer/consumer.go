package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/VsemoGosi-Adventure/events/internal/config"
	"github.com/segmentio/kafka-go"
)

var ErrInvalidMessage = errors.New("invalid crawled event message")

type CrawledEventMessage struct {
	ExternalSource   string      `json:"external_source"`
	ExternalSourceID string      `json:"external_source_id"`
	ExternalURL      string      `json:"external_url"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	StartsAt         time.Time   `json:"starts_at"`
	EndsAt           *time.Time  `json:"ends_at,omitempty"`
	MaxParticipants  *int        `json:"max_participants,omitempty"`
	JoinPolicy       string      `json:"join_policy"`
	Location         LocationDTO `json:"location"`
	Tags             []string    `json:"tags"`
}

type LocationDTO struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

type EventStore interface {
	UpsertCrawledEvent(context.Context, CrawledEventMessage) error
}

type Consumer struct {
	reader *kafka.Reader
	store  EventStore
}

func NewConsumer(cfg config.ConsumerConfig, store EventStore) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfg.Brokers,
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
		store: store,
	}
}

func (c *Consumer) Close() error { return c.reader.Close() }

func (c *Consumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("fetch crawled event failed", "error", err)
			if err := waitRetry(ctx); err != nil {
				return nil
			}
			continue
		}

		for {
			err = c.processMessage(ctx, message.Value)
			if err == nil || errors.Is(err, ErrInvalidMessage) {
				if err != nil {
					slog.Warn("skipping invalid crawled event", "error", err, "partition", message.Partition, "offset", message.Offset)
				}
				break
			}
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("store crawled event failed; retrying", "error", err, "partition", message.Partition, "offset", message.Offset)
			if err := waitRetry(ctx); err != nil {
				return nil
			}
		}

		for {
			if err := c.reader.CommitMessages(ctx, message); err == nil {
				break
			} else if ctx.Err() != nil {
				return nil
			} else {
				slog.Error("commit crawled event offset failed; retrying", "error", err)
			}
			if err := waitRetry(ctx); err != nil {
				return nil
			}
		}
	}
}

func waitRetry(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Consumer) processMessage(ctx context.Context, payload []byte) error {
	var event CrawledEventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}
	if err := event.validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}
	return c.store.UpsertCrawledEvent(ctx, event)
}

func (e CrawledEventMessage) validate() error {
	if strings.TrimSpace(e.ExternalSource) == "" || len(e.ExternalSource) > 255 ||
		strings.TrimSpace(e.ExternalSourceID) == "" || len(e.ExternalSourceID) > 255 ||
		strings.TrimSpace(e.Title) == "" || len(e.Title) > 255 || e.StartsAt.IsZero() {
		return errors.New("source, source ID, title and start time are required")
	}
	if strings.TrimSpace(e.Location.Name) == "" || len(e.Location.Name) > 255 {
		return errors.New("location name is required")
	}
	if e.EndsAt != nil && e.EndsAt.Before(e.StartsAt) {
		return errors.New("end time precedes start time")
	}
	if e.MaxParticipants != nil && *e.MaxParticipants < 2 {
		return errors.New("max participants must be at least two")
	}
	if (e.Location.Latitude == nil) != (e.Location.Longitude == nil) {
		return errors.New("latitude and longitude must both be present")
	}
	if e.Location.Latitude != nil && (math.IsNaN(*e.Location.Latitude) || math.IsInf(*e.Location.Latitude, 0) || *e.Location.Latitude < -90 || *e.Location.Latitude > 90 ||
		math.IsNaN(*e.Location.Longitude) || math.IsInf(*e.Location.Longitude, 0) || *e.Location.Longitude < -180 || *e.Location.Longitude > 180) {
		return errors.New("invalid location coordinates")
	}
	return nil
}
