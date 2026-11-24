package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// ApplySchema creates the tables and indexes of the schema that don't exist yet.
// It doesn't alter existing tables, so a change to an existing table still needs a migration.
func ApplySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}
	return nil
}
