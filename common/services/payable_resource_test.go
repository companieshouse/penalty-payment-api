package services

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/constants"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

var requestId = "123abc5789"
var customerCode = "12345678"
var validPayableRef = "1234"
var invalidPayableRef = "invalid"

func setup(t *testing.T) (*gomock.Controller, *mocks.MockPayableResourceDaoService, *PayableResourceService) {
	mockCtrl := gomock.NewController(t)
	cfg, _ := config.Get()

	mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
	mockPayableResourceSvc := PayableResourceService{
		DAO:    mockPrDaoSvc,
		Config: cfg,
	}

	return mockCtrl, mockPrDaoSvc, &mockPayableResourceSvc
}

func createTestGetRequest(requestId string) *http.Request {
	req := httptest.NewRequest("Get", "/test", nil)
	req.Header.Add("X-Request-ID", requestId)

	return req
}

func buildTestPayableResourceDao(size int, customerCode string, payableRef string, status string) *models.PayableResourceDao {
	transactions := map[string]models.TransactionDao{
		"abcd": {Amount: 5},
	}
	totalAmount := 5
	if size > 1 {
		transactions["wxyz"] = models.TransactionDao{Amount: 10}
		totalAmount = totalAmount + 10
	}
	t := time.Now().Truncate(time.Millisecond)
	return &models.PayableResourceDao{
		CustomerCode: customerCode,
		PayableRef:   payableRef,
		Data: models.PayableResourceDataDao{
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedByDao{
				ID:       "identity",
				Email:    "test@user.com",
				Forename: "some",
				Surname:  "body",
			},
			Links: models.PayableResourceLinksDao{
				Self:    fmt.Sprintf("/company/%s/penalties/payable/%s", customerCode, payableRef),
				Payment: fmt.Sprintf("/company/%s/penalties/payable/%s/payment", customerCode, payableRef),
			},
			Transactions: transactions,
			Payment: models.PaymentDao{
				Status:    status,
				Amount:    strconv.Itoa(totalAmount),
				Reference: "payref",
				PaidAt:    &t,
			},
		},
	}
}

func testGetPayableResource(mockPayableResourceSvc *PayableResourceService, dao *models.PayableResourceDao,
	customerCode string, payableRef string, responseType ResponseType, requestId string) {
	req := createTestGetRequest(requestId)

	payableResource, status, err := mockPayableResourceSvc.GetPayableResource(req, customerCode, payableRef)

	So(status, ShouldEqual, responseType)
	if err != nil {
		So(err.Error(), ShouldEqual, "error getting payable resource from db: [error]")
	} else {
		So(err, ShouldBeNil)
	}
	if payableResource == nil {
		So(payableResource, ShouldBeNil)
	} else {
		So(payableResource.CustomerCode, ShouldEqual, dao.CustomerCode)
		So(payableResource.PayableRef, ShouldEqual, dao.PayableRef)
		So(payableResource.Etag, ShouldEqual, dao.Data.Etag)
		So(payableResource.CreatedAt, ShouldEqual, dao.Data.CreatedAt)
		So(payableResource.CreatedBy.ID, ShouldEqual, dao.Data.CreatedBy.ID)
		So(payableResource.CreatedBy.Email, ShouldEqual, dao.Data.CreatedBy.Email)
		So(payableResource.CreatedBy.Forename, ShouldEqual, dao.Data.CreatedBy.Forename)
		So(payableResource.CreatedBy.Surname, ShouldEqual, dao.Data.CreatedBy.Surname)
		So(payableResource.Links.Self, ShouldEqual, dao.Data.Links.Self)
		So(payableResource.Links.Payment, ShouldEqual, dao.Data.Links.Payment)
		So(payableResource.Payment.Amount, ShouldEqual, dao.Data.Payment.Amount)
		So(payableResource.Payment.Status, ShouldEqual, dao.Data.Payment.Status)
		So(len(payableResource.Transactions), ShouldEqual, len(dao.Data.Transactions))
	}
}

func TestUnitGetPayableResource(t *testing.T) {
	mockCtrl, mockPrDaoSvc, mockPayableResourceSvc := setup(t)

	httpmock.Activate()

	defer httpmock.DeactivateAndReset()
	defer mockCtrl.Finish()

	Convey("Error getting payable resource from DB", t, func() {
		mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(&models.PayableResourceDao{}, fmt.Errorf("error"))

		testGetPayableResource(mockPayableResourceSvc, nil, customerCode, validPayableRef, Error, requestId)
	})

	Convey("Payable resource not found", t, func() {
		mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, invalidPayableRef, requestId).Return(nil, nil)

		testGetPayableResource(mockPayableResourceSvc, nil, customerCode, invalidPayableRef, NotFound, requestId)
	})

	Convey("Get Payable resource - success - Single transaction", t, func() {
		payableResourceDao := buildTestPayableResourceDao(1, customerCode, validPayableRef, "pending")
		mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(
			payableResourceDao, nil,
		)

		testGetPayableResource(mockPayableResourceSvc, payableResourceDao, customerCode, validPayableRef, Success, requestId)
	})

	Convey("Get Payable resource - success - Multiple transactions", t, func() {
		payableResourceDao := buildTestPayableResourceDao(2, customerCode, validPayableRef, "paid")
		mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(
			payableResourceDao,
			nil,
		)

		testGetPayableResource(mockPayableResourceSvc, payableResourceDao, customerCode, validPayableRef, Success, requestId)
	})
}

func buildEmptyPayableResource() models.PayableResource {
	return models.PayableResource{
		CustomerCode: customerCode,
		PayableRef:   validPayableRef,
	}
}

func buildPaymentInformation() validators.PaymentInformation {
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	completedAt, _ := time.Parse(layout, str)

	return validators.PaymentInformation{
		Reference:   "123",
		Amount:      "150",
		Status:      "paid",
		CompletedAt: completedAt,
		CreatedBy:   "test@example.com",
	}
}

func TestUnitPayableResourceService_UpdateAsPaid(t *testing.T) {
	Convey("PayableResourceService.UpdateAsPaid", t, func() {
		mockCtrl, mockPrDaoSvc, mockPayableResourceSvc := setup(t)

		defer mockCtrl.Finish()

		Convey("Payable resource must exist", func() {
			mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(nil, errors.New("not found"))

			err := mockPayableResourceSvc.UpdateAsPaid(buildEmptyPayableResource(), validators.PaymentInformation{}, requestId)

			So(err, ShouldBeError, ErrPenaltyNotFound)
		})

		Convey("Penalty payable resource must not have already been paid", func() {
			payableResourceDao := buildTestPayableResourceDao(1, customerCode, validPayableRef, "paid")
			mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(payableResourceDao, nil)

			err := mockPayableResourceSvc.UpdateAsPaid(buildEmptyPayableResource(), validators.PaymentInformation{Status: constants.Paid.String()}, requestId)

			So(err, ShouldBeError, ErrAlreadyPaid)
		})

		Convey("payment details are saved to db", func() {
			payableResourceDao := buildTestPayableResourceDao(1, customerCode, validPayableRef, "pending")
			mockPrDaoSvc.EXPECT().GetPayableResource(customerCode, validPayableRef, requestId).Return(payableResourceDao, nil)
			mockPrDaoSvc.EXPECT().UpdatePaymentDetails(payableResourceDao, requestId).Times(1)

			paymentResponse := buildPaymentInformation()

			err := mockPayableResourceSvc.UpdateAsPaid(buildEmptyPayableResource(), paymentResponse, requestId)

			So(err, ShouldBeNil)
			So(payableResourceDao.Data.Payment.Status, ShouldEqual, paymentResponse.Status)
			So(payableResourceDao.Data.Payment.PaidAt, ShouldNotBeNil)
			So(payableResourceDao.Data.Payment.Amount, ShouldEqual, paymentResponse.Amount)
			So(payableResourceDao.Data.Payment.Reference, ShouldEqual, paymentResponse.Reference)
		})
	})
}
