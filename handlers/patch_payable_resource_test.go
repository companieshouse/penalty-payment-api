package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"
	"github.com/companieshouse/go-session-handler/httpsession"
	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/penalty-payment-api-core/constants"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
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
var penaltyDetailsMap = &config.PenaltyDetailsMap{}
var allowedTransactionsMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
			"S1": true,
		},
	},
}

// reduces the boilerplate code needed to create, dispatch and unmarshal response body
func dispatchPayResourceHandler(ctx context.Context, t *testing.T, reqBody *models.PatchResourceRequest,
	daoSvc dao.PayableResourceDaoService) (*httptest.ResponseRecorder, *models.ResponseResource) {

	payableResourceService := &services.PayableResourceService{}

	if daoSvc != nil {
		payableResourceService.DAO = daoSvc
	}

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatal("failed to marshal request body")
		}
		body = bytes.NewReader(b)
	}

	ctx = context.WithValue(ctx, httpsession.ContextKeySession, &session.Session{})

	h := PayResourceHandler(payableResourceService, e5.NewClient("foo", "e5api"), penaltyDetailsMap, allowedTransactionsMap)
	req := httptest.NewRequest(http.MethodPost, "/", body).WithContext(ctx)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req.WithContext(ctx))

	if res.Body.Len() > 0 {
		var responseBody models.ResponseResource
		err := json.NewDecoder(res.Body).Decode(&responseBody)
		if err != nil {
			t.Errorf("failed to read response body")
		}

		return res, &responseBody
	}

	return res, nil
}

// Mock function for erroring when preparing and sending kafka message
func mockSendEmailKafkaMessageError(payableResource models.PayableResource, req *http.Request,
	detailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) error {
	return errors.New("error")
}

// Mock function for successful preparing and sending of kafka message
func mockSendEmailKafkaMessage(payableResource models.PayableResource, req *http.Request,
	detailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) error {
	return nil
}

func TestUnitPayResourceHandler(t *testing.T) {
	Convey("PayResourceHandler tests", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("payable resource must be in context", func() {
			res, body := dispatchPayResourceHandler(context.Background(), t, nil, nil)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
			So(body.Message, ShouldEqual, "no payable request present in request context")
		})

		Convey("payment reference is required in request body", func() {
			ctx := context.WithValue(context.Background(), config.PayableResource, &models.PayableResource{})
			res, body := dispatchPayResourceHandler(ctx, t, &models.PatchResourceRequest{}, nil)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
			So(body.Message, ShouldEqual, "the request contained insufficient data and/or failed validation")
		})

		Convey("bad responses from payment api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(
				http.MethodGet,
				"/payments/123",
				httpmock.NewStringResponder(404, ""),
			)

			model := &models.PayableResource{PayableRef: "123"}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)
			reqBody := &models.PatchResourceRequest{Reference: "123"}

			res, body := dispatchPayResourceHandler(ctx, t, reqBody, nil)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
			So(body.Message, ShouldEqual, "the payable resource does not exist")
		})

		Convey("payment (from payments api) is not paid", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// stub the response from the payments api
			p := &companieshouseapi.PaymentResource{Status: "failed", Amount: "150"}
			responder, _ := httpmock.NewJsonResponder(http.StatusOK, p)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/payments/123",
				responder,
			)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/private/payments/123/payment-details",
				httpmock.NewStringResponder(http.StatusOK, "{}"),
			)

			// the payable resource in the request context
			model := &models.PayableResource{PayableRef: "123"}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)

			reqBody := &models.PatchResourceRequest{Reference: "123"}
			res, body := dispatchPayResourceHandler(ctx, t, reqBody, nil)

			So(res.Code, ShouldEqual, http.StatusBadRequest)
			So(body.Message, ShouldEqual, "there was a problem validating this payment")
		})

		Convey("problem with sending confirmation email", func() {
			mockedGetCompanyCode := func(penaltyReference string) (string, error) {
				return utils.LateFilingPenalty, nil
			}

			getCompanyCode = mockedGetCompanyCode

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// stub the response from the payments api
			p := &companieshouseapi.PaymentResource{Status: "paid", Amount: "0", Reference: "financial_penalty_123"}
			responder, _ := httpmock.NewJsonResponder(http.StatusOK, p)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/payments/123",
				responder,
			)

			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/private/payments/123/payment-details",
				httpmock.NewStringResponder(http.StatusOK, "{}"),
			)

			// stub the mongo lookup
			dataModel := &models.PayableResourceDao{}
			mockService := mocks.NewMockService(mockCtrl)
			mockService.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			mockService.EXPECT().UpdatePaymentDetails(dataModel).Times(1)
			mockService.EXPECT().SaveE5Error("", "123", e5.CreateAction).Return(errors.New(""))

			// the payable resource in the request context
			model := &models.PayableResource{
				PayableRef: "123",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "A1234567"},
				},
			}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)

			// stub kafka message
			handleEmailKafkaMessage = mockSendEmailKafkaMessageError

			reqBody := &models.PatchResourceRequest{Reference: "123"}
			res, body := dispatchPayResourceHandler(ctx, t, reqBody, mockService)

			So(dataModel.IsPaid(), ShouldBeTrue)
			So(res.Code, ShouldEqual, http.StatusInternalServerError)
			So(body, ShouldBeNil)
		})

		Convey("Penalty has already been paid", func() {
			mockedGetCompanyCode := func(penaltyReference string) (string, error) {
				return utils.LateFilingPenalty, nil
			}

			getCompanyCode = mockedGetCompanyCode

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// stub the response from the payments api
			p := &companieshouseapi.PaymentResource{Status: "paid", Amount: "0", Reference: "financial_penalty_123"}
			responder, _ := httpmock.NewJsonResponder(http.StatusOK, p)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/payments/123",
				responder,
			)

			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/private/payments/123/payment-details",
				httpmock.NewStringResponder(http.StatusOK, "{}"),
			)

			// stub the mongo lookup
			dataModel := &models.PayableResourceDao{
				Data: models.PayableResourceDataDao{
					Payment: models.PaymentDao{
						Status: constants.Paid.String(),
					},
				},
			}

			mockService := mocks.NewMockService(mockCtrl)
			mockService.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			mockService.EXPECT().SaveE5Error("", "123", e5.CreateAction).Return(errors.New(""))

			// the payable resource in the request context
			model := &models.PayableResource{
				PayableRef: "123",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "A1234567"},
				},
			}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)

			// stub kafka message
			handleEmailKafkaMessage = mockSendEmailKafkaMessage

			reqBody := &models.PatchResourceRequest{Reference: "123"}
			res, body := dispatchPayResourceHandler(ctx, t, reqBody, mockService)

			So(dataModel.IsPaid(), ShouldBeTrue)
			So(res.Code, ShouldEqual, http.StatusInternalServerError)
			So(body, ShouldBeNil)
		})

		Convey("problem with sending request to E5", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// stub the response from the payments api
			p := &companieshouseapi.PaymentResource{Status: "paid", Amount: "0", Reference: "financial_penalty_123"}
			responder, _ := httpmock.NewJsonResponder(http.StatusOK, p)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/payments/123",
				responder,
			)

			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/private/payments/123/payment-details",
				httpmock.NewStringResponder(http.StatusOK, "{}"),
			)

			// stub the response from the e5 api
			e5Responder := httpmock.NewStringResponder(http.StatusBadRequest, e5ValidationError)
			httpmock.RegisterResponder(http.MethodPost, "/arTransactions/payment", e5Responder)

			// stub the mongo lookup
			dataModel := &models.PayableResourceDao{}
			mockService := mocks.NewMockService(mockCtrl)
			mockService.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			mockService.EXPECT().UpdatePaymentDetails(dataModel).Times(1)
			mockService.EXPECT().SaveE5Error("", "123", e5.CreateAction).Return(errors.New(""))

			// the payable resource in the request context
			model := &models.PayableResource{
				PayableRef: "123",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "A1234567"},
				},
			}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)

			// stub kafka message
			handleEmailKafkaMessage = mockSendEmailKafkaMessage

			reqBody := &models.PatchResourceRequest{Reference: "123"}
			res, body := dispatchPayResourceHandler(ctx, t, reqBody, mockService)

			So(dataModel.IsPaid(), ShouldBeTrue)
			So(res.Code, ShouldEqual, http.StatusInternalServerError)
			So(body, ShouldBeNil)
		})

		Convey("success when payment is valid", func() {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			// stub the response from the payments api
			p := &companieshouseapi.PaymentResource{
				Status:    "paid",
				Amount:    "150",
				Reference: "financial_penalty_123",
				CreatedBy: companieshouseapi.CreatedBy{
					Email: "test@example.com",
				},
			}

			responder, _ := httpmock.NewJsonResponder(http.StatusOK, p)
			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/payments/123",
				responder,
			)

			httpmock.RegisterResponder(
				http.MethodGet,
				companieshouseapi.PaymentsBasePath+"/private/payments/123/payment-details",
				httpmock.NewStringResponder(http.StatusOK, "{}"),
			)

			// stub the response from the e5 api
			e5Responder := httpmock.NewBytesResponder(http.StatusOK, nil)
			httpmock.RegisterResponder(http.MethodPost, "e5api/arTransactions/payment", e5Responder)
			httpmock.RegisterResponder(http.MethodPost, "e5api/arTransactions/payment/authorise", e5Responder)
			httpmock.RegisterResponder(http.MethodPost, "e5api/arTransactions/payment/confirm", e5Responder)

			// stub the mongo lookup
			dataModel := &models.PayableResourceDao{}
			mockService := mocks.NewMockService(mockCtrl)
			mockService.EXPECT().GetPayableResource(gomock.Any(), gomock.Any()).Return(dataModel, nil)
			mockService.EXPECT().UpdatePaymentDetails(dataModel).Times(1)

			// the payable resource in the request context
			model := &models.PayableResource{
				PayableRef:   "123",
				CustomerCode: "10000024",
				Transactions: []models.TransactionItem{
					{PenaltyRef: "A1234567", Amount: 150},
				},
			}
			ctx := context.WithValue(context.Background(), config.PayableResource, model)

			// stub kafka message
			handleEmailKafkaMessage = mockSendEmailKafkaMessage

			reqBody := &models.PatchResourceRequest{Reference: "123"}
			res, body := dispatchPayResourceHandler(ctx, t, reqBody, mockService)

			So(res.Code, ShouldEqual, http.StatusNoContent)
			So(body, ShouldBeNil)
		})
	})
}
