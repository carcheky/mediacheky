package database

import (
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitialize(t *testing.T) {
	log := logger.New("debug")

	cfg := Config{
		Path: ":memory:",
	}

	db, err := Initialize(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify database is working
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestRunMigrations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = RunMigrations(db)
	assert.NoError(t, err)

	// Verify all tables exist
	tables := []string{"services", "templates", "global_config", "service_logs", "media", "schedules", "history", "settings"}
	for _, table := range tables {
		var tableName string
		err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName).Error
		assert.NoError(t, err)
		assert.Equal(t, table, tableName, "Table %s should exist", table)
	}
}

func TestSeedData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations first
	err = RunMigrations(db)
	require.NoError(t, err)

	// Seed data
	err = SeedData(db)
	assert.NoError(t, err)

	// Verify global config was seeded
	var configCount int64
	db.Model(&models.GlobalConfig{}).Count(&configCount)
	assert.Greater(t, configCount, int64(0))

	// Verify specific config exists
	var puid models.GlobalConfig
	err = db.Where("key = ?", "PUID").First(&puid).Error
	assert.NoError(t, err)
	assert.Equal(t, "1000", puid.Value)
	assert.Equal(t, "system", puid.Category)

	// Verify templates were seeded
	var templateCount int64
	db.Model(&models.Template{}).Count(&templateCount)
	assert.Greater(t, templateCount, int64(0))

	// Verify specific template exists
	var radarrTemplate models.Template
	err = db.Where("name = ?", "radarr").First(&radarrTemplate).Error
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", radarrTemplate.Version)
	assert.Contains(t, radarrTemplate.Content, "linuxserver/radarr")
}

func TestSeedData_Idempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = RunMigrations(db)
	require.NoError(t, err)

	// Seed data twice
	err = SeedData(db)
	require.NoError(t, err)
	err = SeedData(db)
	require.NoError(t, err)

	// Verify no duplicates
	var configCount int64
	db.Model(&models.GlobalConfig{}).Count(&configCount)

	var templateCount int64
	db.Model(&models.Template{}).Count(&templateCount)

	// Should have exactly the seeded amount, not double
	assert.Equal(t, int64(8), configCount)   // 8 default configs
	assert.Equal(t, int64(3), templateCount) // 3 default templates
}

func TestClose(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = Close(db)
	assert.NoError(t, err)
}
