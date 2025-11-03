package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRepository_Set(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	err := repo.Set("PUID", "1000", "system")
	assert.NoError(t, err)

	// Retrieve and verify
	config, err := repo.GetByKey("PUID")
	assert.NoError(t, err)
	assert.Equal(t, "PUID", config.Key)
	assert.Equal(t, "1000", config.Value)
	assert.Equal(t, "system", config.Category)
}

func TestConfigRepository_Set_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Set initial value
	err := repo.Set("TZ", "UTC", "system")
	require.NoError(t, err)

	// Update value
	err = repo.Set("TZ", "America/New_York", "system")
	assert.NoError(t, err)

	// Verify update
	config, err := repo.GetByKey("TZ")
	assert.NoError(t, err)
	assert.Equal(t, "America/New_York", config.Value)

	// Ensure only one record exists
	all, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestConfigRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Create multiple config entries
	configs := map[string]struct {
		value    string
		category string
	}{
		"PUID":      {"1000", "system"},
		"PGID":      {"1000", "system"},
		"TZ":        {"UTC", "system"},
		"BASE_PATH": {"/data", "paths"},
	}

	for key, cfg := range configs {
		err := repo.Set(key, cfg.value, cfg.category)
		require.NoError(t, err)
	}

	// Get all configs
	allConfigs, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allConfigs, 4)
}

func TestConfigRepository_GetByCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Create configs in different categories
	err := repo.Set("PUID", "1000", "system")
	require.NoError(t, err)
	err = repo.Set("PGID", "1000", "system")
	require.NoError(t, err)
	err = repo.Set("BASE_PATH", "/data", "paths")
	require.NoError(t, err)
	err = repo.Set("NETWORK_MODE", "bridge", "network")
	require.NoError(t, err)

	// Get system configs
	systemConfigs, err := repo.GetByCategory("system")
	assert.NoError(t, err)
	assert.Len(t, systemConfigs, 2)

	// Verify they are system configs
	for _, cfg := range systemConfigs {
		assert.Equal(t, "system", cfg.Category)
	}
}

func TestConfigRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Create a config
	err := repo.Set("TEST_KEY", "test_value", "test")
	require.NoError(t, err)

	// Delete it
	err = repo.Delete("TEST_KEY")
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByKey("TEST_KEY")
	assert.Error(t, err)
}

func TestConfigRepository_GetAsMap(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Create configs
	configs := map[string]string{
		"PUID":      "1000",
		"PGID":      "1000",
		"TZ":        "UTC",
		"BASE_PATH": "/data",
	}

	for key, value := range configs {
		err := repo.Set(key, value, "system")
		require.NoError(t, err)
	}

	// Get as map
	configMap, err := repo.GetAsMap()
	assert.NoError(t, err)
	assert.Len(t, configMap, 4)

	// Verify values
	for key, expectedValue := range configs {
		assert.Equal(t, expectedValue, configMap[key])
	}
}

func TestConfigRepository_NonExistentKey(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepository(db)

	// Try to get non-existent key
	_, err := repo.GetByKey("NON_EXISTENT")
	assert.Error(t, err)
}
