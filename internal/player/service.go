package player

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mhetem/DH-Companion/internal/db"
	"github.com/mhetem/DH-Companion/internal/srd"
)

type Service struct {
	ctx     context.Context
	q       *db.Queries
	conn    *sql.DB
	catalog *srd.Catalog
}

func New() *Service { return &Service{} }

func Attach(s *Service, ctx context.Context, q *db.Queries, conn *sql.DB, catalog *srd.Catalog) {
	s.ctx, s.q, s.conn, s.catalog = ctx, q, conn, catalog
}

func (s *Service) tx(fn func(*db.Queries) error) error {
	tx, err := s.conn.BeginTx(s.ctx, nil)
	if err != nil {
		return fmt.Errorf("opening transaction: %w", err)
	}
	defer tx.Rollback()
	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

var ErrNotFound = errors.New("not found")

func notFound(kind, key string) error {
	return fmt.Errorf("%s %q %w", kind, key, ErrNotFound)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
