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
	penaltyPayment  = newPenaltyPayment(1)
	penaltyPayment2 = newPenaltyPayment(2)
	penaltyPayment3 = newPenaltyPayment(3)
	e5PaymentID     = "XKIYLUq1pRVuiLNA"
	cfg             = &config.Config{
		PenaltyPaymentsProcessingTopic:         "penalty-payments-processing",
		PenaltyPaymentsProcessingMaxRetries:    "3",
		PenaltyPaymentsProcessingRetryDelay:    "1",
		PenaltyPaymentsProcessingRetryMaxDelay: "5",
		ConsumerGroupName:                      "penalty-payment-api-penalty-payments-processing",
		ConsumerRetryGroupName:                 "penalty-payment-api-penalty-payments-processing-retry",
		ConsumerRetryThrottleRate:              1,
		ConsumerRetryMaxAttempts:               3,
		FeatureFlagPaymentsProcessingEnabled:   true,
	}
)

type mockE5Client struct {
	mock.Mock
}

func (m *mockE5Client) GetTransactions(input *e5.GetTransactionsInput, _ string) (*e5.GetTransactionsResponse, error) {
	m.Called(input)
	return nil, errors.New("get transactions not used")
}

func (m *mockE5Client) TimeoutPayment(input *e5.PaymentActionInput, _ string) error {
	m.Called(input)
	return errors.New("timeout payment not used")
}

func (m *mockE5Client) RejectPayment(input *e5.PaymentActionInput, _ string) error {
	m.Called(input)
	return errors.New("reject payment not used")
}

func (m *mockE5Client) CreatePayment(input *e5.CreatePaymentInput, _ string) error {
	return m.Called(input).Error(0)
}

func (m *mockE5Client) AuthorisePayment(input *e5.AuthorisePaymentInput, _ string) error {
	return m.Called(input).Error(0)
}

func (m *mockE5Client) ConfirmPayment(input *e5.PaymentActionInput, _ string) error {
	return m.Called(input).Error(0)
}

var _ e5.ClientInterface = (*mockE5Client)(nil)

type mockDAO struct {
	mock.Mock
}

func (m *mockDAO) CreatePayableResource(dao *models.PayableResourceDao, _ string) error {
	m.Called(dao)
	return errors.New("create payable resource not used")
}

func (m *mockDAO) GetPayableResource(customerCode, payableRef, _ string) (*models.PayableResourceDao, error) {
	m.Called(customerCode, payableRef)
	return nil, errors.New("get payable resource not used")
}

func (m *mockDAO) UpdatePaymentDetails(dao *models.PayableResourceDao, _ string) error {
	m.Called(dao)
	return errors.New("update payment details not used")
}

func (m *mockDAO) Shutdown() {
	panic("shutdown not used")
}

func (m *mockDAO) SaveE5Error(customerCode, payableRef, _ string, action e5.Action) error {
	return m.Called(customerCode, payableRef, action).Error(0)
}

func TestUnitProcessFinancialPenaltyPayment_IsAfter24Hours(t *testing.T) {
	Convey("Process financial penalty payment is after 24 hours", t, func() {
		// Given
		e5Client, DAO := setupMocks()
		handler := &PenaltyFinancePayment{
			E5Client:                  e5Client,
			PayableResourceDaoService: DAO,
		}

		penaltyPaymentToSkip := models.PenaltyPaymentsProcessing{
			CreatedAt: "2025-08-19T08:52:11.648+00:00",
		}

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPaymentToSkip, e5PaymentID, cfg, false)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
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
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg, false)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertNotCalled(t, "SaveE5Error", mock.Anything)
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

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg, false)

		// Then
		So(err, ShouldBeError, errors.New("All attempts fail:\n#1: create payment in E5 failed\n#2: create payment in E5 failed\n#3: create payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertNotCalled(t, "SaveE5Error", penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5.CreateAction)
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
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg, false)

		// Then
		So(err, ShouldBeNil)
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
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg, false)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_Retry_Success(t *testing.T) {
	// Given
	e5Client, DAO := setupMocks()
	handler := &PenaltyFinancePayment{
		E5Client:                  e5Client,
		PayableResourceDaoService: DAO,
	}

	e5Client.On("CreatePayment", mock.Anything).Return(nil)
	e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
	e5Client.On("ConfirmPayment", mock.Anything).Return(nil)

	Convey("Process financial penalty payment retry success with Attempt = 2", t, func() {
		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment2, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertNotCalled(t, "SaveE5Error", mock.Anything)
	})

	Convey("Process financial penalty payment retry success with Attempt = 3", t, func() {
		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment3, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertNotCalled(t, "SaveE5Error", mock.Anything)
	})
}

func TestUnitProcessFinancialPenaltyPayment_Retry_CreatePaymentFails(t *testing.T) {
	// Given
	e5Client, DAO := setupMocks()
	handler := &PenaltyFinancePayment{
		E5Client:                  e5Client,
		PayableResourceDaoService: DAO,
	}

	e5Client.On("CreatePayment", mock.Anything).Return(errors.New("create payment in E5 failed"))

	Convey("Process financial penalty payment retry create payment fails with Attempt = 2", t, func() {
		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment2, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeError, errors.New("All attempts fail:\n#1: create payment in E5 failed\n#2: create payment in E5 failed\n#3: create payment in E5 failed"))
		e5Client.AssertExpectations(t)
		DAO.AssertNotCalled(t, "SaveE5Error", penaltyPayment2.CustomerCode, penaltyPayment2.PayableRef, e5.CreateAction)
	})

	Convey("Process financial penalty payment retry create payment fails with Attempt = 3", t, func() {
		DAO.On("SaveE5Error", penaltyPayment3.CustomerCode, penaltyPayment3.PayableRef, e5.CreateAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment3, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_Retry_AuthorisePaymentFails(t *testing.T) {
	// Given
	e5Client, DAO := setupMocks()
	handler := &PenaltyFinancePayment{
		E5Client:                  e5Client,
		PayableResourceDaoService: DAO,
	}

	e5Client.On("CreatePayment", mock.Anything).Return(nil)
	e5Client.On("AuthorisePayment", mock.Anything).Return(errors.New("authorise payment in E5 failed"))

	Convey("Process financial penalty payment retry authorise payment fails with Attempt = 2", t, func() {
		DAO.On("SaveE5Error", penaltyPayment2.CustomerCode, penaltyPayment2.PayableRef, e5.AuthoriseAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment2, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})

	Convey("Process financial penalty payment retry authorise payment fails with Attempt = 3", t, func() {
		DAO.On("SaveE5Error", penaltyPayment3.CustomerCode, penaltyPayment3.PayableRef, e5.AuthoriseAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment3, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})
}

func TestUnitProcessFinancialPenaltyPayment_Retry_ConfirmPaymentFails(t *testing.T) {
	// Given
	e5Client, DAO := setupMocks()
	handler := &PenaltyFinancePayment{
		E5Client:                  e5Client,
		PayableResourceDaoService: DAO,
	}

	e5Client.On("CreatePayment", mock.Anything).Return(nil)
	e5Client.On("AuthorisePayment", mock.Anything).Return(nil)
	e5Client.On("ConfirmPayment", mock.Anything).Return(errors.New("confirm payment in E5 failed"))

	Convey("Process financial penalty payment retry confirm payment fails with Attempt = 2", t, func() {
		DAO.On("SaveE5Error", penaltyPayment2.CustomerCode, penaltyPayment2.PayableRef, e5.ConfirmAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment2, e5PaymentID, cfg, true)

		// Then
		So(err, ShouldBeNil)
		e5Client.AssertExpectations(t)
		DAO.AssertExpectations(t)
	})

	Convey("Process financial penalty payment retry confirm payment fails with Attempt = 3", t, func() {
		DAO.On("SaveE5Error", penaltyPayment3.CustomerCode, penaltyPayment3.PayableRef, e5.ConfirmAction).Return(nil)

		// When
		err := handler.ProcessFinancialPenaltyPayment(penaltyPayment3, e5PaymentID, cfg, true)

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

func newPenaltyPayment(attempt int32) models.PenaltyPaymentsProcessing {
	return models.PenaltyPaymentsProcessing{
		Attempt:           attempt,
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
}
