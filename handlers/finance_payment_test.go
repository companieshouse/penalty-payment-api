package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

var (
	penaltyPayment = models.PenaltyPaymentsProcessing{
		Attempt:           1,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		CompanyCode:       "C1",
		CustomerCode:      "OE123456",
		PaymentID:         "KIYLUq1pRVuiLNA",
		ExternalPaymentID: "a8n3vp4uo1o7mf7pp2mtab7ne9",
		PaymentReference:  "financial_penalty_SQ33133143",
		PaymentAmount:     "350.00",
		TotalValue:        350.0,
		TransactionPayments: []models.TransactionPayment{{
			TransactionReference: "U1234567",
			Value:                350.0,
		}},
		CardType:   "Visa",
		Email:      "test@example.com",
		PayableRef: "SQ33133143",
	}
	e5PaymentID = "XKIYLUq1pRVuiLNA"
	cfg         = &config.Config{}
)

type mockE5Client struct {
	mock.Mock
}

func (m *mockE5Client) GetTransactions(input *e5.GetTransactionsInput) (*e5.GetTransactionsResponse, error) {
	m.Called(input)
	return nil, errors.New("get transactions not used")
}

func (m *mockE5Client) TimeoutPayment(input *e5.PaymentActionInput) error {
	m.Called(input)
	return errors.New("timeout payment not used")
}

func (m *mockE5Client) RejectPayment(input *e5.PaymentActionInput) error {
	m.Called(input)
	return errors.New("reject payment not used")
}

func (m *mockE5Client) CreatePayment(input *e5.CreatePaymentInput) error {
	return m.Called(input).Error(0)
}

func (m *mockE5Client) AuthorisePayment(input *e5.AuthorisePaymentInput) error {
	return m.Called(input).Error(0)
}

func (m *mockE5Client) ConfirmPayment(input *e5.PaymentActionInput) error {
	return m.Called(input).Error(0)
}

var _ e5.ClientInterface = (*mockE5Client)(nil)

type mockDAO struct {
	mock.Mock
}

func (m *mockDAO) CreatePayableResource(dao *models.PayableResourceDao) error {
	m.Called(dao)
	return errors.New("create payable resource not used")
}

func (m *mockDAO) GetPayableResource(customerCode, payableRef string) (*models.PayableResourceDao, error) {
	m.Called(customerCode, payableRef)
	return nil, errors.New("get payable resource not used")
}

func (m *mockDAO) UpdatePaymentDetails(dao *models.PayableResourceDao) error {
	m.Called(dao)
	return errors.New("update payment details not used")
}

func (m *mockDAO) Shutdown() {
	panic("shutdown not used")
}

func (m *mockDAO) SaveE5Error(customerCode, payableRef string, action e5.Action) error {
	return m.Called(customerCode, payableRef, action).Error(0)
}

func TestUnitProcessFinancialPenaltyPayment_Success(t *testing.T) {
	Convey("Process financial penalty payment success", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
		e5Client.On("ConfirmPayment", mock.Anything).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_CreatePaymentFails(t *testing.T) {
	Convey("Process financial penalty payment create payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(errors.New("create payment in E5 failed"))
		DAO.On("SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.CreateAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg)

		// Then
		So(err, ShouldBeError, errors.New("All attempts fail:\n#1: create payment in E5 failed\n#2: create payment in E5 failed\n#3: create payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_AuthorisePaymentFails(t *testing.T) {
	Convey("Process financial penalty payment authorise payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(errors.New("authorise payment in E5 failed"))
		DAO.On("SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.AuthoriseAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg)

		// Then
		So(err, ShouldBeError, errors.New("All attempts fail:\n#1: authorise payment in E5 failed\n#2: authorise payment in E5 failed\n#3: authorise payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_ConfirmPaymentFails(t *testing.T) {
	Convey("Process financial penalty payment confirm payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
		e5Client.On("ConfirmPayment", mock.Anything).Return(errors.New("confirm payment in E5 failed"))
		DAO.On("SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.ConfirmAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg)

		// Then
		So(err, ShouldBeError, errors.New("All attempts fail:\n#1: confirm payment in E5 failed\n#2: confirm payment in E5 failed\n#3: confirm payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

// retry
func TestUnitProcessFinancialPenaltyPaymentRetryTopic_Success(t *testing.T) {
	Convey("Process financial penalty payment success", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
		e5Client.On("ConfirmPayment", mock.Anything).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPaymentRetryTopic(penaltyPayment, e5PaymentID)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPaymentRetryTopic_CreatePaymentFails(t *testing.T) {
	Convey("Process financial penalty payment create payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(errors.New("create payment in E5 failed"))

		// When
		err := handler.ProcessFinancialPenaltyPaymentRetryTopic(penaltyPayment, e5PaymentID)

		// Then
		So(err, ShouldBeError, errors.New("create payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPaymentRetryTopic_AuthorisePaymentFails(t *testing.T) {
	Convey("Process financial penalty payment authorise payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(errors.New("authorise payment in E5 failed"))
		DAO.On("SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.AuthoriseAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPaymentRetryTopic(penaltyPayment, e5PaymentID)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPaymentRetryTopic_ConfirmPaymentFails(t *testing.T) {
	Convey("Process financial penalty payment confirm payment fails", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		e5Client.On("CreatePayment", mock.Anything).Return(nil)
		e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
		e5Client.On("ConfirmPayment", mock.Anything).Return(errors.New("confirm payment in E5 failed"))
		DAO.On("SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.ConfirmAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPaymentRetryTopic(penaltyPayment, e5PaymentID)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func setupMocks() (*mockE5Client, *mockDAO) {
	e5Client := new(mockE5Client)
	DAO := new(mockDAO)
	return e5Client, DAO
}
