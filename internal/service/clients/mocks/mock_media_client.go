package mocks

import (
	"context"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockMediaClient is a mock implementation of the MediaClient interface
type MockMediaClient struct {
	mock.Mock
}

// TestConnection mocks the TestConnection method
func (m *MockMediaClient) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// GetLibrary mocks the GetLibrary method
func (m *MockMediaClient) GetLibrary(ctx context.Context) ([]*models.Media, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}

// GetItem mocks the GetItem method
func (m *MockMediaClient) GetItem(ctx context.Context, id int) (*models.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Media), args.Error(1)
}

// DeleteItem mocks the DeleteItem method
func (m *MockMediaClient) DeleteItem(ctx context.Context, id int, deleteFiles bool) error {
	args := m.Called(ctx, id, deleteFiles)
	return args.Error(0)
}

// GetTags mocks the GetTags method
func (m *MockMediaClient) GetTags(ctx context.Context) ([]models.Tag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Tag), args.Error(1)
}
