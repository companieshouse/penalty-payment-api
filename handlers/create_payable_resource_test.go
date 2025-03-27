package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var companyNumber = "10000024"

func serveCreatePayableResourceHandler(body []byte, service dao.Service, withAuthUserDetails bool) *httptest.ResponseRecorder {
	template := "/company/%s/penalties/payable"
	path := fmt.Sprintf(template, companyNumber)
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

	ctx = context.WithValue(ctx, config.CustomerCode, companyNumber)
	return ctx
}

var e5ResponseLateFiling = `
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
      "transactionReference": "A1234567",
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

var e5ResponseSanctions = `
{
  "page": {
    "size": 4,
    "totalElements": 4,
    "totalPages": 1,
    "number": 0
  },
  "data": [
    {
      "companyCode": "C1",
      "ledgerCode": "EW",
      "customerCode": "10000024",
      "transactionReference": "P1234567",
      "transactionDate": "2017-11-28",
      "madeUpDate": "2017-02-28",
      "amount": 150,
      "outstandingAmount": 150,
      "isPaid": false,
      "transactionType": "1",
      "transactionSubType": "S1",
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
      "transactionReference": "A1234567",
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
      "transactionReference": "A0378421",
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

	url := "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" + utils.LateFilingPenalty + "&fromDate=1990-01-01"

	Convey("Error decoding request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := []byte{'{'}
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when user details not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company number not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{})
		companyNumber = ""
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)
		companyNumber = "10000024"

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company code cannot be determined", t, func() {
		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "", errors.New("no penalty reference found")
		}
		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{})
		companyNumber = ""
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)
		companyNumber = "10000024"

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need at least one transaction", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenalty)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
		})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Only allowed 1 transaction in a resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenalty)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseMultipleTx))
		mockService := mocks.NewMockService(mockCtrl)

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{TransactionID: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
				{TransactionID: "A0378421", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("internal server error when failing to create payable resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenalty)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))
		mockService := mocks.NewMockService(mockCtrl)

		// expect the CreatePayableResource to be called once and return an error
		mockService.EXPECT().CreatePayableResource(gomock.Any()).Return(errors.New("any error"))

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{TransactionID: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("successfully creating a payable request", t, func() {
		testCases := []struct {
			name             string
			companyCode      string
			penaltyReference string
			urlE5            string
			e5Response       string
		}{
			{
				name:             "Late Filing",
				companyCode:      utils.LateFilingPenalty,
				penaltyReference: "A1234567",
				urlE5: "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" +
					utils.LateFilingPenalty + "&fromDate=1990-01-01",
				e5Response: e5ResponseLateFiling,
			},
			{
				name:             "Sanctions",
				companyCode:      utils.Sanctions,
				penaltyReference: "P1234567",
				urlE5: "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" +
					utils.Sanctions + "&fromDate=1990-01-01",
				e5Response: e5ResponseSanctions,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				setGetCompanyCodeFromTransactionMock(tc.companyCode)

				httpmock.Activate()
				mockCtrl := gomock.NewController(t)
				defer httpmock.DeactivateAndReset()
				defer mockCtrl.Finish()

				httpmock.RegisterResponder("GET", tc.urlE5, httpmock.NewStringResponder(200, tc.e5Response))
				mockService := mocks.NewMockService(mockCtrl)

				// expect the CreatePayableResource to be called once and return without error
				mockService.EXPECT().CreatePayableResource(gomock.Any()).Return(nil)

				body, _ := json.Marshal(&models.PayableRequest{
					CustomerCode: "10000024",
					CreatedBy:    authentication.AuthUserDetails{},
					Transactions: []models.TransactionItem{
						{TransactionID: tc.penaltyReference, Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
					},
				})

				res := serveCreatePayableResourceHandler(body, mockService, true)

				So(res.Code, ShouldEqual, http.StatusCreated)
				So(res.Header().Get("Content-Type"), ShouldEqual, "application/json")
			})
		}
	})
}
