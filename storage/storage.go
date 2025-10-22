package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hoenirvili/axiogate/log"
)

type Storage struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

type Option func(p *Storage)

func WithLogger(log *slog.Logger) Option {
	return func(s *Storage) {
		s.log = log
	}
}

func New(db *pgxpool.Pool, options ...Option) *Storage {
	s := &Storage{
		db:  db,
		log: log.Noop(),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func (r *Storage) Save(ctx context.Context, provider string, payload []byte) error {
	query := `INSERT INTO shipment (provider, payload) VALUES ($1, $2)`
	r.log.With(
		slog.String("query", query),
		slog.Group("record",
			slog.String("provider", provider),
			slog.String("payload", string(payload)),
		)).Debug("Save")
	if _, err := r.db.Exec(ctx, query, provider, payload); err != nil {
		return fmt.Errorf("failed to save shipment, %w", err)
	}
	return nil
}
