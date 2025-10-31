package mocks

import (
	"context"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockStreamingClient is a mock implementation of the StreamingClient interface
type MockStreamingClient struct {
	mock.Mock
}

// TestConnection mocks the TestConnection method
func (m *MockStreamingClient) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// GetLibrary mocks the GetLibrary method
func (m *MockStreamingClient) GetLibrary(ctx context.Context) ([]*models.Media, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}

// GetPlaybackInfo mocks the GetPlaybackInfo method
func (m *MockStreamingClient) GetPlaybackInfo(ctx context.Context, mediaID string) (*models.PlaybackInfo, error) {
	args := m.Called(ctx, mediaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlaybackInfo), args.Error(1)
}

// DeleteItem mocks the DeleteItem method
func (m *MockStreamingClient) DeleteItem(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
