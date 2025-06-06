package services

import (
	"errors"
	"fmt"
	"net/http/httptest"
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

func createMockPayableResourceService(mockDAO *mocks.MockPayableResourceDaoService, cfg *config.Config) PayableResourceService {
	return PayableResourceService{
		DAO:    mockDAO,
		Config: cfg,
	}
}

func TestUnitGetPayableResource(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	cfg, _ := config.Get()

	Convey("Error getting payable resource from DB", t, func() {
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", gomock.Any()).Return(&models.PayableResourceDao{}, fmt.Errorf("error"))

		req := httptest.NewRequest("Get", "/test", nil)

		payableResource, status, err := mockPayableResourceSvc.GetPayableResource(req, "12345678", "1234")
		So(payableResource, ShouldBeNil)
		So(status, ShouldEqual, Error)
		So(err.Error(), ShouldEqual, "error getting payable resource from db: [error]")
	})

	Convey("Payable resource not found", t, func() {
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "invalid").Return(nil, nil)

		req := httptest.NewRequest("Get", "/test", nil)

		payableResource, status, err := mockPayableResourceSvc.GetPayableResource(req, "12345678", "invalid")
		So(payableResource, ShouldBeNil)
		So(status, ShouldEqual, NotFound)
		So(err, ShouldBeNil)
	})

	Convey("Get Payable resource - success - Single transaction", t, func() {
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		t := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", gomock.Any()).Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
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
						Self:    "/company/12345678/penalties/payable/1234",
						Payment: "/company/12345678/penalties/payable/1234/payment",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		req := httptest.NewRequest("Get", "/test", nil)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		payableResource, status, err := mockPayableResourceSvc.GetPayableResource(req, "12345678", "1234")

		So(status, ShouldEqual, Success)
		So(err, ShouldBeNil)
		So(payableResource.CustomerCode, ShouldEqual, "12345678")
		So(payableResource.PayableRef, ShouldEqual, "1234")
		So(payableResource.Etag, ShouldEqual, "qwertyetag1234")
		So(payableResource.CreatedAt, ShouldEqual, &t)
		So(payableResource.CreatedBy.ID, ShouldEqual, "identity")
		So(payableResource.CreatedBy.Email, ShouldEqual, "test@user.com")
		So(payableResource.CreatedBy.Forename, ShouldEqual, "some")
		So(payableResource.CreatedBy.Surname, ShouldEqual, "body")
		So(payableResource.Links.Self, ShouldEqual, "/company/12345678/penalties/payable/1234")
		So(payableResource.Links.Payment, ShouldEqual, "/company/12345678/penalties/payable/1234/payment")
		So(payableResource.Payment.Amount, ShouldEqual, "5")
		So(payableResource.Payment.Status, ShouldEqual, "pending")
		So(len(payableResource.Transactions), ShouldEqual, 1)
		So(payableResource.Transactions[0].Amount, ShouldEqual, 5)
	})

	Convey("Get Payable resource - success - Multiple transactions", t, func() {
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
			"wxyz": {Amount: 10},
		}
		t := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", gomock.Any()).Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
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
						Self:    "/company/12345678/penalties/payable/1234",
						Payment: "/company/12345678/penalties/payable/1234/payment",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status:    constants.Paid.String(),
						Amount:    "15",
						Reference: "payref",
						PaidAt:    &t,
					},
				},
			},
			nil,
		)

		req := httptest.NewRequest("Get", "/test", nil)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		payableResource, status, err := mockPayableResourceSvc.GetPayableResource(req, "12345678", "1234")

		So(status, ShouldEqual, Success)
		So(err, ShouldBeNil)
		So(payableResource.CustomerCode, ShouldEqual, "12345678")
		So(payableResource.PayableRef, ShouldEqual, "1234")
		So(payableResource.Etag, ShouldEqual, "qwertyetag1234")
		So(payableResource.CreatedAt, ShouldEqual, &t)
		So(payableResource.CreatedBy.ID, ShouldEqual, "identity")
		So(payableResource.CreatedBy.Email, ShouldEqual, "test@user.com")
		So(payableResource.CreatedBy.Forename, ShouldEqual, "some")
		So(payableResource.CreatedBy.Surname, ShouldEqual, "body")
		So(payableResource.Links.Self, ShouldEqual, "/company/12345678/penalties/payable/1234")
		So(payableResource.Links.Payment, ShouldEqual, "/company/12345678/penalties/payable/1234/payment")
		So(payableResource.Payment.Amount, ShouldEqual, "15")
		So(payableResource.Payment.Status, ShouldEqual, "paid")
		So(payableResource.Payment.Reference, ShouldEqual, "payref")
		So(payableResource.Payment.PaidAt, ShouldEqual, &t)
		So(len(payableResource.Transactions), ShouldEqual, 2)
		So(payableResource.Transactions[0].Amount+payableResource.Transactions[1].Amount, ShouldEqual, 15) // array order can change - sum can't
	})
}

func TestUnitPayableResourceService_UpdateAsPaid(t *testing.T) {
	Convey("PayableResourceService.UpdateAsPaid", t, func() {
		Convey("Payable resource must exist", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
			mockPrDaoSvc.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
			svc := PayableResourceService{DAO: mockPrDaoSvc}

			err := svc.UpdateAsPaid(models.PayableResource{}, validators.PaymentInformation{})

			So(err, ShouldBeError, ErrPenaltyNotFound)
		})

		Convey("Penalty payable resource must not have already been paid", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			dataModel := &models.PayableResourceDao{
				Data: models.PayableResourceDataDao{
					Payment: models.PaymentDao{
						Status: constants.Paid.String(),
					},
				},
			}
			mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
			mockPrDaoSvc.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			svc := PayableResourceService{DAO: mockPrDaoSvc}

			err := svc.UpdateAsPaid(models.PayableResource{}, validators.PaymentInformation{Status: constants.Paid.String()})

			So(err, ShouldBeError, ErrAlreadyPaid)
		})

		Convey("payment details are saved to db", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			dataModel := &models.PayableResourceDao{
				Data: models.PayableResourceDataDao{
					Payment: models.PaymentDao{},
				},
			}
			mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
			mockPrDaoSvc.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			mockPrDaoSvc.EXPECT().UpdatePaymentDetails(gomock.Any()).Times(1)
			svc := PayableResourceService{DAO: mockPrDaoSvc}

			layout := "2006-01-02T15:04:05.000Z"
			str := "2014-11-12T11:45:26.371Z"
			completedAt, _ := time.Parse(layout, str)

			paymentResponse := validators.PaymentInformation{
				Reference:   "123",
				Amount:      "150",
				Status:      "paid",
				CompletedAt: completedAt,
				CreatedBy:   "test@example.com",
			}

			err := svc.UpdateAsPaid(models.PayableResource{}, paymentResponse)

			So(err, ShouldBeNil)
			So(dataModel.Data.Payment.Status, ShouldEqual, paymentResponse.Status)
			So(dataModel.Data.Payment.PaidAt, ShouldNotBeNil)
			So(dataModel.Data.Payment.Amount, ShouldEqual, paymentResponse.Amount)
			So(dataModel.Data.Payment.Reference, ShouldEqual, paymentResponse.Reference)
		})
	})
}
