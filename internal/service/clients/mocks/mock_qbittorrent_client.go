package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockQBittorrentClient is a mock implementation of the QBittorrentClient
type MockQBittorrentClient struct {
	mock.Mock
}

// DeleteTorrent mocks the DeleteTorrent method
func (m *MockQBittorrentClient) DeleteTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	args := m.Called(ctx, hash, deleteFiles)
	return args.Error(0)
}

// TestConnection mocks the TestConnection method
func (m *MockQBittorrentClient) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
