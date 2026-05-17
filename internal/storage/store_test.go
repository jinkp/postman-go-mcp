package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempDB(t *testing.T) (*storage.SQLiteStore, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := storage.New(path)
	require.NoError(t, err)
	return store, func() {
		store.Close()
		os.Remove(path)
	}
}

func TestSaveScanAndList(t *testing.T) {
	store, cleanup := tempDB(t)
	defer cleanup()

	err := store.SaveScan("./api/openapi.yaml", 42)
	require.NoError(t, err)

	records, err := store.ListScans()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "./api/openapi.yaml", records[0].Source)
	assert.Equal(t, 42, records[0].EndpointCount)
}

func TestSaveCollection(t *testing.T) {
	store, cleanup := tempDB(t)
	defer cleanup()

	err := store.SaveCollection("My API", "./api/openapi.yaml", "./postman/collection.json")
	require.NoError(t, err)
}

func TestListScansEmpty(t *testing.T) {
	store, cleanup := tempDB(t)
	defer cleanup()

	records, err := store.ListScans()
	require.NoError(t, err)
	assert.Empty(t, records)
}
