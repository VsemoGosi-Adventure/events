package storage

import (
	"github.com/VsemoGosi-Adventure/events/internal/repo/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool:    pool,
		Queries: db.New(pool),
	}
}

func (s *Storage) Close() error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}
