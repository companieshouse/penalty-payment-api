package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockE5Service is a mock implementation of the E5 service
type MockE5Service struct {
	mock.Mock
}

func (m *MockE5Service) PerformAction(action string) error {
	args := m.Called(action)
	return args.Error(0)
}