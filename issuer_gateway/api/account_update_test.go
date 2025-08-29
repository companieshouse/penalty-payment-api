package api

import (
	j "encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

var e5ValidationError = `

	{
	 "httpStatusCode" : 400,
	 "status" : "BAD_REQUEST",
	 "timestamp" : "2019-07-07T18:40:07Z",
	 "messageCode" : null,
	 "message" : "Constraint Validation error",
	 "debugMessage" : null,
	 "subErrors" : [ {
	   "object" : "String",
	   "field" : "companyCode",
	   "rejectedValue" : "LPs",
	   "message" : "size must be between 0 and 2"
	 } ]
	}

`

func TestUnitUpdateIssuerAccountWithPenaltyPaid(t *testing.T) {
	Convey("amount must be okay to parse as float", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		payableResourceSvc := &services.PayableResourceService{DAO: mockPrDaoSvc}

		c := &e5.Client{}
		r := models.PayableResource{}
		p := validators.PaymentInformation{Amount: "foo"}

		err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")
		So(err, ShouldNotBeNil)
	})

	Convey("invalid company code", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockPayableResourceDaoService(mockCtrl)
		svc := &services.PayableResourceService{DAO: mockService}

		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "", errors.New("cannot determine company code")
		}
		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		c := &e5.Client{}
		p := validators.PaymentInformation{Amount: "150", PaymentID: "123"}
		r := models.PayableResource{
			PayableRef:   "123",
			CustomerCode: "10000024",
			Transactions: []models.TransactionItem{
				{PenaltyRef: "123", Amount: 150},
			},
		}

		err := UpdateIssuerAccountWithPenaltyPaid(svc, c, r, p, "")

		So(err, ShouldBeError, "cannot determine company code")
	})

	Convey("E5 request errors", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		payableResourceSvc := &services.PayableResourceService{DAO: mockPrDaoSvc}

		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return utils.LateFilingPenaltyCompanyCode, nil
		}
		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		Convey("failure in creating a new payment", func() {
			defer httpmock.Reset()
			e5Responder := httpmock.NewStringResponder(http.StatusBadRequest, e5ValidationError)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", e5Responder)

			mockPrDaoSvc.EXPECT().SaveE5Error("10000024", "123", e5.CreateAction).Return(errors.New(""))

			c := &e5.Client{}
			p := validators.PaymentInformation{Amount: "150", PaymentID: "123"}
			r := models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "123", Amount: 150},
				},
			}

			err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")

			So(err, ShouldBeError, e5.ErrE5BadRequest)
		})

		Convey("failure in authorising a payment", func() {
			defer httpmock.Reset()
			e5Responder := httpmock.NewStringResponder(http.StatusBadRequest, e5ValidationError)
			okResponder := httpmock.NewBytesResponder(http.StatusOK, nil)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", okResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/authorise", e5Responder)

			mockPrDaoSvc.EXPECT().SaveE5Error("10000024", "123", e5.AuthoriseAction).Return(errors.New(""))

			c := &e5.Client{}
			p := validators.PaymentInformation{
				Amount:    "150",
				PaymentID: "123",
				CreatedBy: "test@example.com",
			}

			r := models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "123", Amount: 150},
				},
			}

			err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")

			So(err, ShouldBeError, e5.ErrE5BadRequest)
		})

		Convey("failure in confirming a payment", func() {
			defer httpmock.Reset()
			e5Responder := httpmock.NewStringResponder(http.StatusBadRequest, e5ValidationError)
			okResponder := httpmock.NewBytesResponder(http.StatusOK, nil)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", okResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/authorise", okResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/confirm", e5Responder)

			mockPrDaoSvc.EXPECT().SaveE5Error("10000024", "123", e5.ConfirmAction).Return(errors.New(""))

			c := &e5.Client{}
			p := validators.PaymentInformation{
				Amount:    "150",
				PaymentID: "123",
				CreatedBy: "test@example.com",
			}

			r := models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "123", Amount: 150},
				},
			}

			err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")

			So(err, ShouldBeError, e5.ErrE5BadRequest)
		})

		Convey("no errors when all 3 calls to E5 succeed", func() {
			defer httpmock.Reset()
			okResponder := httpmock.NewBytesResponder(http.StatusOK, nil)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", okResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/authorise", okResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/confirm", okResponder)

			c := &e5.Client{}
			p := validators.PaymentInformation{
				Amount:    "150",
				PaymentID: "123",
				CreatedBy: "test@example.com",
			}

			r := models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "123", Amount: 150},
				},
			}

			err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")

			So(err, ShouldBeNil)
		})

		Convey("paymentId (PUON) is prefixed with 'X'", func() {
			defer httpmock.Reset()

			// struct to decode the request body
			type body struct {
				PaymentID string `json:"paymentId"`
			}

			// check the payment id value before responding.
			paymentIDResponder := func(req *http.Request) (*http.Response, error) {
				resp := httpmock.NewBytesResponse(http.StatusOK, nil)
				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(req.Body)
				var b body
				err := j.NewDecoder(req.Body).Decode(&b)
				if err != nil {
					return nil, errors.New("failed to read request body")
				}

				if b.PaymentID[0] != 'X' {
					return nil, errors.New("paymentId does not begin with an X")
				}
				return resp, nil
			}

			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", paymentIDResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/authorise", paymentIDResponder)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment/confirm", paymentIDResponder)

			c := &e5.Client{}
			p := validators.PaymentInformation{
				Amount:    "150",
				PaymentID: "123",
				CreatedBy: "test@example.com",
			}

			r := models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "123", Amount: 150},
				},
			}

			err := UpdateIssuerAccountWithPenaltyPaid(payableResourceSvc, c, r, p, "")
			So(err, ShouldBeNil)

		})
	})
}
