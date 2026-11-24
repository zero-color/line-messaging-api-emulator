package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplySchema(t *testing.T) {
	ctx := context.Background()
	pool, err := ConnectDB(testConnectionString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	const schemaName = "apply_schema_test"
	_, err = pool.Exec(ctx, "CREATE SCHEMA "+schemaName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DROP SCHEMA "+schemaName+" CASCADE")
	})

	config := pool.Config()
	config.ConnConfig.RuntimeParams["search_path"] = schemaName
	schemaPool, err := newPool(ctx, config)
	require.NoError(t, err)
	t.Cleanup(schemaPool.Close)

	countTables := func(t *testing.T) int {
		t.Helper()
		var count int
		err := schemaPool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1", schemaName).Scan(&count)
		require.NoError(t, err)
		return count
	}

	t.Run("creates every table in an empty database", func(t *testing.T) {
		require.NoError(t, ApplySchema(ctx, schemaPool))

		assert.Equal(t, 11, countTables(t))
	})

	t.Run("applying again changes nothing", func(t *testing.T) {
		require.NoError(t, ApplySchema(ctx, schemaPool))

		assert.Equal(t, 11, countTables(t))
	})

	t.Run("creates the tables missing from an older database", func(t *testing.T) {
		_, err := schemaPool.Exec(ctx, "DROP TABLE rich_menus CASCADE")
		require.NoError(t, err)

		require.NoError(t, ApplySchema(ctx, schemaPool))

		assert.Equal(t, 11, countTables(t))
	})
}
