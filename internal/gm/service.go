package gm

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

var ErrNotFound = errors.New("not found")

func notFound(kind, key string) error {
	return fmt.Errorf("%s %q %w", kind, key, ErrNotFound)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

type Filter struct {
	Tier string `json:"tier"`
	Type string `json:"type"`
}

func (f Filter) normalized() (tier, typ string) {
	tier = strings.TrimSpace(f.Tier)
	typ = strings.TrimSpace(f.Type)
	if typ == "" {
		typ = "All"
	}
	return tier, typ
}

func nullString(p *string) sql.NullString {
	if p == nil || *p == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func stringPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	v := n.String
	return &v
}

func nullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func int64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}
