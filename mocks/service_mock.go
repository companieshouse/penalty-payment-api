// Code generated by MockGen. DO NOT EDIT.
// Source: dao/service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	"github.com/companieshouse/lfp-pay-api-core/models"
	"github.com/companieshouse/penalty-payment-api/e5"
	"github.com/golang/mock/gomock"
	"reflect"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// CreatePayableResource mocks base method
func (m *MockService) CreatePayableResource(dao *models.PayableResourceDao) error {
	ret := m.ctrl.Call(m, "CreatePayableResource", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePayableResource indicates an expected call of CreatePayableResource
func (mr *MockServiceMockRecorder) CreatePayableResource(dao interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePayableResource", reflect.TypeOf((*MockService)(nil).CreatePayableResource), dao)
}

// GetPayableResource mocks base method
func (m *MockService) GetPayableResource(companyNumber, reference string) (*models.PayableResourceDao, error) {
	ret := m.ctrl.Call(m, "GetPayableResource", companyNumber, reference)
	ret0, _ := ret[0].(*models.PayableResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayableResource indicates an expected call of GetPayableResource
func (mr *MockServiceMockRecorder) GetPayableResource(companyNumber, reference interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayableResource", reflect.TypeOf((*MockService)(nil).GetPayableResource), companyNumber, reference)
}

// UpdatePaymentDetails mocks base method
func (m *MockService) UpdatePaymentDetails(dao *models.PayableResourceDao) error {
	ret := m.ctrl.Call(m, "UpdatePaymentDetails", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePaymentDetails indicates an expected call of UpdatePaymentDetails
func (mr *MockServiceMockRecorder) UpdatePaymentDetails(dao interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePaymentDetails", reflect.TypeOf((*MockService)(nil).UpdatePaymentDetails), dao)
}

// SaveE5Error mocks base method
func (m *MockService) SaveE5Error(companyNumber, reference string, action e5.Action) error {
	ret := m.ctrl.Call(m, "SaveE5Error", companyNumber, reference, action)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveE5Error indicates an expected call of SaveE5Error
func (mr *MockServiceMockRecorder) SaveE5Error(companyNumber, reference, action interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveE5Error", reflect.TypeOf((*MockService)(nil).SaveE5Error), companyNumber, reference, action)
}

// Shutdown mocks base method
func (m *MockService) Shutdown() {
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown
func (mr *MockServiceMockRecorder) Shutdown() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockService)(nil).Shutdown))
}
