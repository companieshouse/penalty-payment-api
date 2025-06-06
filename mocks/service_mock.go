// Code generated by MockGen. DO NOT EDIT.

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/companieshouse/penalty-payment-api-core/models"
	e5 "github.com/companieshouse/penalty-payment-api/common/e5"
	gomock "github.com/golang/mock/gomock"
)

// MockPayableResourceDaoService is a mock of PayableResourceDaoService interface.
type MockPayableResourceDaoService struct {
	ctrl     *gomock.Controller
	recorder *MockPayableResourceDaoServiceMockRecorder
}

// MockPayableResourceDaoServiceMockRecorder is the mock recorder for MockPayableResourceDaoService.
type MockPayableResourceDaoServiceMockRecorder struct {
	mock *MockPayableResourceDaoService
}

// NewMockPayableResourceDaoService creates a new mock instance.
func NewMockPayableResourceDaoService(ctrl *gomock.Controller) *MockPayableResourceDaoService {
	mock := &MockPayableResourceDaoService{ctrl: ctrl}
	mock.recorder = &MockPayableResourceDaoServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPayableResourceDaoService) EXPECT() *MockPayableResourceDaoServiceMockRecorder {
	return m.recorder
}

// CreatePayableResource mocks base method.
func (m *MockPayableResourceDaoService) CreatePayableResource(dao *models.PayableResourceDao) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePayableResource", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePayableResource indicates an expected call of CreatePayableResource.
func (mr *MockPayableResourceDaoServiceMockRecorder) CreatePayableResource(dao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePayableResource", reflect.TypeOf((*MockPayableResourceDaoService)(nil).CreatePayableResource), dao)
}

// GetPayableResource mocks base method.
func (m *MockPayableResourceDaoService) GetPayableResource(customerCode, payableRef string) (*models.PayableResourceDao, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPayableResource", customerCode, payableRef)
	ret0, _ := ret[0].(*models.PayableResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPayableResource indicates an expected call of GetPayableResource.
func (mr *MockPayableResourceDaoServiceMockRecorder) GetPayableResource(customerCode, payableRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPayableResource", reflect.TypeOf((*MockPayableResourceDaoService)(nil).GetPayableResource), customerCode, payableRef)
}

// SaveE5Error mocks base method.
func (m *MockPayableResourceDaoService) SaveE5Error(customerCode, payableRef string, action e5.Action) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveE5Error", customerCode, payableRef, action)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveE5Error indicates an expected call of SaveE5Error.
func (mr *MockPayableResourceDaoServiceMockRecorder) SaveE5Error(customerCode, payableRef, action interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveE5Error", reflect.TypeOf((*MockPayableResourceDaoService)(nil).SaveE5Error), customerCode, payableRef, action)
}

// Shutdown mocks base method.
func (m *MockPayableResourceDaoService) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockPayableResourceDaoServiceMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockPayableResourceDaoService)(nil).Shutdown))
}

// UpdatePaymentDetails mocks base method.
func (m *MockPayableResourceDaoService) UpdatePaymentDetails(dao *models.PayableResourceDao) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePaymentDetails", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePaymentDetails indicates an expected call of UpdatePaymentDetails.
func (mr *MockPayableResourceDaoServiceMockRecorder) UpdatePaymentDetails(dao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePaymentDetails", reflect.TypeOf((*MockPayableResourceDaoService)(nil).UpdatePaymentDetails), dao)
}

// MockAccountPenaltiesDaoService is a mock of AccountPenaltiesDaoService interface.
type MockAccountPenaltiesDaoService struct {
	ctrl     *gomock.Controller
	recorder *MockAccountPenaltiesDaoServiceMockRecorder
}

// MockAccountPenaltiesDaoServiceMockRecorder is the mock recorder for MockAccountPenaltiesDaoService.
type MockAccountPenaltiesDaoServiceMockRecorder struct {
	mock *MockAccountPenaltiesDaoService
}

// NewMockAccountPenaltiesDaoService creates a new mock instance.
func NewMockAccountPenaltiesDaoService(ctrl *gomock.Controller) *MockAccountPenaltiesDaoService {
	mock := &MockAccountPenaltiesDaoService{ctrl: ctrl}
	mock.recorder = &MockAccountPenaltiesDaoServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountPenaltiesDaoService) EXPECT() *MockAccountPenaltiesDaoServiceMockRecorder {
	return m.recorder
}

// CreateAccountPenalties mocks base method.
func (m *MockAccountPenaltiesDaoService) CreateAccountPenalties(dao *models.AccountPenaltiesDao) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccountPenalties", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAccountPenalties indicates an expected call of CreateAccountPenalties.
func (mr *MockAccountPenaltiesDaoServiceMockRecorder) CreateAccountPenalties(dao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccountPenalties", reflect.TypeOf((*MockAccountPenaltiesDaoService)(nil).CreateAccountPenalties), dao)
}

// GetAccountPenalties mocks base method.
func (m *MockAccountPenaltiesDaoService) GetAccountPenalties(customerCode, companyCode string) (*models.AccountPenaltiesDao, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountPenalties", customerCode, companyCode)
	ret0, _ := ret[0].(*models.AccountPenaltiesDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountPenalties indicates an expected call of GetAccountPenalties.
func (mr *MockAccountPenaltiesDaoServiceMockRecorder) GetAccountPenalties(customerCode, companyCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountPenalties", reflect.TypeOf((*MockAccountPenaltiesDaoService)(nil).GetAccountPenalties), customerCode, companyCode)
}

// UpdateAccountPenalties mocks base method.
func (m *MockAccountPenaltiesDaoService) UpdateAccountPenalties(dao *models.AccountPenaltiesDao) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountPenalties", dao)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountPenalties indicates an expected call of UpdateAccountPenalties.
func (mr *MockAccountPenaltiesDaoServiceMockRecorder) UpdateAccountPenalties(dao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountPenalties", reflect.TypeOf((*MockAccountPenaltiesDaoService)(nil).UpdateAccountPenalties), dao)
}

// UpdateAccountPenaltyAsPaid mocks base method.
func (m *MockAccountPenaltiesDaoService) UpdateAccountPenaltyAsPaid(customerCode, companyCode, penaltyRef string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountPenaltyAsPaid", customerCode, companyCode, penaltyRef)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountPenaltyAsPaid indicates an expected call of UpdateAccountPenaltyAsPaid.
func (mr *MockAccountPenaltiesDaoServiceMockRecorder) UpdateAccountPenaltyAsPaid(customerCode, companyCode, penaltyRef interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountPenaltyAsPaid", reflect.TypeOf((*MockAccountPenaltiesDaoService)(nil).UpdateAccountPenaltyAsPaid), customerCode, companyCode, penaltyRef)
}
