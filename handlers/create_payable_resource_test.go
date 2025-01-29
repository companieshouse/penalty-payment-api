package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/dao"
	"github.com/companieshouse/penalty-payment-api/middleware"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

var companyNumber = "10000024"
var penaltyNumber = "A1234567"

func serveCreatePayableResourceHandler(body []byte, service dao.Service, withAuthUserDetails bool) *httptest.ResponseRecorder {
	template := "/company/%s/penalties/late-filing/%s/payable"
	path := fmt.Sprintf(template, companyNumber, penaltyNumber)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler := CreatePayableResourceHandler(service, penaltyDetailsMap, allowedTransactionsMap)
	handler.ServeHTTP(res, req.WithContext(testContext(withAuthUserDetails)))

	return res
}

func testContext(withAuthUserDetails bool) context.Context {
	ctx := context.Background()
	if withAuthUserDetails {
		ctx = context.WithValue(ctx, authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
	} else {
		ctx = context.WithValue(ctx, authentication.ContextKeyUserDetails, nil)
	}

	details := middleware.CompanyDetails{M: map[string]string{
		"CompanyNumber": companyNumber,
	}}

	ctx = context.WithValue(ctx, config.CompanyDetails, details)
	return ctx
}

var e5Response = `
{
  "page": {
    "size": 4,
    "totalElements": 4,
    "totalPages": 1,
    "number": 0
  },
  "data": [
    {
      "companyCode": "LP",
      "ledgerCode": "EW",
      "customerCode": "10000024",
      "transactionReference": "00378420",
      "transactionDate": "2017-11-28",
      "madeUpDate": "2017-02-28",
      "amount": 150,
      "outstandingAmount": 150,
      "isPaid": false,
      "transactionType": "1",
      "transactionSubType": "EU",
      "typeDescription": "Penalty Ltd Wel & Eng <=1m     LTDWA    ",
      "dueDate": "2017-12-12"
    }
  ]
}
`

var e5ResponseMultipleTx = `
{
  "page": {
    "size": 4,
    "totalElements": 4,
    "totalPages": 1,
    "number": 0
  },
  "data": [
    {
      "companyCode": "LP",
      "ledgerCode": "EW",
      "customerCode": "10000024",
      "transactionReference": "00378420",
      "transactionDate": "2017-11-28",
      "madeUpDate": "2017-02-28",
      "amount": 150,
      "outstandingAmount": 150,
      "isPaid": false,
      "transactionType": "1",
      "transactionSubType": "EU",
      "typeDescription": "Penalty Ltd Wel & Eng <=1m     LTDWA    ",
      "dueDate": "2017-12-12"
    },
    {
      "companyCode": "LP",
      "ledgerCode": "EW",
      "customerCode": "10000024",
      "transactionReference": "00378421",
      "transactionDate": "2017-11-28",
      "madeUpDate": "2017-02-28",
      "amount": 150,
      "outstandingAmount": 150,
      "isPaid": false,
      "transactionType": "1",
      "transactionSubType": "EU",
      "typeDescription": "Penalty Ltd Wel & Eng <=1m     LTDWA    ",
      "dueDate": "2017-12-12"
    }
  ]
}
`

func TestUnitCreatePayableResourceHandler(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		return
	}
	cfg, _ := config.Get()
	cfg.E5APIURL = "https://e5"
	cfg.E5Username = "SYSTEM"

	url := "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=LP&fromDate=1990-01-01"

	Convey("Error decoding request body", t, func() {
		mockedGetCompanyCode := func(companyNumber string) (string, error) {
			return "LP", nil
		}
		getCompanyCode = mockedGetCompanyCode

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))

		body := []byte{'{'}
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when user details not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))

		body, _ := json.Marshal(&models.PayableRequest{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when company number not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))

		body, _ := json.Marshal(&models.PayableRequest{})
		companyNumber = ""
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)
		companyNumber = "10000024"

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need at least one transaction", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))

		body, _ := json.Marshal(&models.PayableRequest{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Only allowed 1 transaction in a resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseMultipleTx))
		mockService := mocks.NewMockService(mockCtrl)

		body, _ := json.Marshal(&models.PayableRequest{
			CompanyNumber: "10000024",
			CreatedBy:     authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{TransactionID: "00378420", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
				{TransactionID: "00378421", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("internal server error when failing to create payable resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))
		mockService := mocks.NewMockService(mockCtrl)

		// expect the CreatePayableResource to be called once and return an error
		mockService.EXPECT().CreatePayableResource(gomock.Any()).Return(errors.New("any error"))

		body, _ := json.Marshal(&models.PayableRequest{
			CompanyNumber: "10000024",
			CreatedBy:     authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{TransactionID: "00378420", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("successfully creating a payable request", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5Response))
		mockService := mocks.NewMockService(mockCtrl)

		// expect the CreatePayableResource to be called once and return without error
		mockService.EXPECT().CreatePayableResource(gomock.Any()).Return(nil)

		body, _ := json.Marshal(&models.PayableRequest{
			CompanyNumber: "10000024",
			CreatedBy:     authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{TransactionID: "00378420", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
		So(res.Header().Get("Content-Type"), ShouldEqual, "application/json")
	})
}
