package repository

import (
	"testing"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	log := &models.ServiceLog{
		ServiceID:   1,
		ServiceName: "radarr",
		Action:      "start",
		Status:      "success",
		Message:     "Service started successfully",
	}

	err := repo.Create(log)
	assert.NoError(t, err)
	assert.NotZero(t, log.ID)
	assert.NotZero(t, log.CreatedAt)
}

func TestServiceLogRepository_GetByServiceID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	// Create logs for different services
	logs := []*models.ServiceLog{
		{ServiceID: 1, ServiceName: "radarr", Action: "start", Status: "success"},
		{ServiceID: 1, ServiceName: "radarr", Action: "stop", Status: "success"},
		{ServiceID: 2, ServiceName: "sonarr", Action: "start", Status: "failed"},
	}

	for _, log := range logs {
		err := repo.Create(log)
		require.NoError(t, err)
		// Sleep briefly to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Get logs for service 1
	serviceLogs, err := repo.GetByServiceID(1, 10)
	assert.NoError(t, err)
	assert.Len(t, serviceLogs, 2)

	// Verify they are ordered by created_at DESC
	assert.Equal(t, "stop", serviceLogs[0].Action)
	assert.Equal(t, "start", serviceLogs[1].Action)
}

func TestServiceLogRepository_GetRecent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	// Create multiple logs
	logs := []*models.ServiceLog{
		{ServiceID: 1, ServiceName: "radarr", Action: "start", Status: "success"},
		{ServiceID: 2, ServiceName: "sonarr", Action: "start", Status: "success"},
		{ServiceID: 1, ServiceName: "radarr", Action: "restart", Status: "success"},
	}

	for _, log := range logs {
		err := repo.Create(log)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	// Get recent logs
	recentLogs, err := repo.GetRecent(2)
	assert.NoError(t, err)
	assert.Len(t, recentLogs, 2)

	// Verify ordering (most recent first)
	assert.Equal(t, "restart", recentLogs[0].Action)
	assert.Equal(t, "start", recentLogs[1].Action)
}

func TestServiceLogRepository_GetByAction(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	// Create logs with different actions
	logs := []*models.ServiceLog{
		{ServiceID: 1, ServiceName: "radarr", Action: "start", Status: "success"},
		{ServiceID: 1, ServiceName: "radarr", Action: "stop", Status: "success"},
		{ServiceID: 2, ServiceName: "sonarr", Action: "start", Status: "failed"},
		{ServiceID: 1, ServiceName: "radarr", Action: "restart", Status: "success"},
	}

	for _, log := range logs {
		err := repo.Create(log)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	// Get logs by action "start"
	startLogs, err := repo.GetByAction("start", 10)
	assert.NoError(t, err)
	assert.Len(t, startLogs, 2)

	for _, log := range startLogs {
		assert.Equal(t, "start", log.Action)
	}
}

func TestServiceLogRepository_DeleteOlderThan(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	// Create old log (simulate by manually setting CreatedAt)
	oldLog := &models.ServiceLog{
		ServiceID:   1,
		ServiceName: "radarr",
		Action:      "start",
		Status:      "success",
	}
	err := repo.Create(oldLog)
	require.NoError(t, err)

	// Manually update CreatedAt to be 10 days old
	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	db.Model(&models.ServiceLog{}).Where("id = ?", oldLog.ID).Update("created_at", tenDaysAgo)

	// Create recent log
	recentLog := &models.ServiceLog{
		ServiceID:   1,
		ServiceName: "radarr",
		Action:      "stop",
		Status:      "success",
	}
	err = repo.Create(recentLog)
	require.NoError(t, err)

	// Delete logs older than 7 days
	deleted, err := repo.DeleteOlderThan(7)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// Verify only recent log remains
	allLogs, err := repo.GetRecent(10)
	assert.NoError(t, err)
	assert.Len(t, allLogs, 1)
	assert.Equal(t, recentLog.ID, allLogs[0].ID)
}

func TestServiceLogRepository_Limit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceLogRepository(db)

	// Create 5 logs
	for i := 0; i < 5; i++ {
		log := &models.ServiceLog{
			ServiceID:   1,
			ServiceName: "radarr",
			Action:      "start",
			Status:      "success",
		}
		err := repo.Create(log)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	// Get with limit
	logs, err := repo.GetByServiceID(1, 3)
	assert.NoError(t, err)
	assert.Len(t, logs, 3)
}
