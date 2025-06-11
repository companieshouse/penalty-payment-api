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

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var customerCode = "10000024"

func serveCreatePayableResourceHandler(body []byte, payableResourceService dao.PayableResourceDaoService, apDaoSvc dao.AccountPenaltiesDaoService,
	withAuthUserDetails bool) *httptest.ResponseRecorder {
	template := "/company/%s/penalties/payable"
	path := fmt.Sprintf(template, customerCode)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler := CreatePayableResourceHandler(payableResourceService, apDaoSvc, penaltyDetailsMap, allowedTransactionsMap)
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

	ctx = context.WithValue(ctx, config.CustomerCode, customerCode)
	return ctx
}

var e5ResponseLateFiling = `
{
  "page": {
    "size": 1,
    "totalElements": 1,
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
      "dueDate": "2017-12-12",
      "accountStatus": "CHS",
      "dunningStatus": "PEN1        "
    }
  ]
}
`

var e5ResponseSanctions = `
{
  "page": {
    "size": 1,
    "totalElements": 1,
    "totalPages": 1,
    "number": 0
  },
  "data": [
    {
      "companyCode": "C1",
      "ledgerCode": "E1",
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
      "dueDate": "2017-12-12",
      "accountStatus": "CHS",
      "dunningStatus": "PEN1        "
    }
  ]
}
`

var e5ResponseSanctionsRoe = `
{
  "page": {
    "size": 1,
    "totalElements": 1,
    "totalPages": 1,
    "number": 0
  },
  "data": [
    {
      "companyCode": "C1",
      "ledgerCode": "FU",
      "customerCode": "OE123456",
      "transactionReference": "U1234567",
      "transactionDate": "2017-11-28",
      "madeUpDate": "2017-02-28",
      "amount": 150,
      "outstandingAmount": 150,
      "isPaid": false,
      "transactionType": "1",
      "transactionSubType": "A2",
      "typeDescription": "Failure to update",
      "dueDate": "2017-12-12",
      "accountStatus": "CHS",
      "dunningStatus": "PEN1        "
    }
  ]
}
`

var e5ResponseMultipleTx = `
{
  "page": {
    "size": 2,
    "totalElements": 2,
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
      "dueDate": "2017-12-12",
      "accountStatus": "CHS",
      "dunningStatus": "PEN1        "
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
      "dueDate": "2017-12-12",
      "accountStatus": "CHS",
      "dunningStatus": "PEN1        "
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

	url := "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" + utils.LateFilingPenaltyCompanyCode + "&fromDate=1990-01-01"

	Convey("Error decoding request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := []byte{'{'}
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when user details not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company number not in context", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{})
		customerCode = ""
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true)
		customerCode = "10000024"

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company code cannot be resolved", t, func() {
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
		customerCode = ""
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true)
		customerCode = "10000024"

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need at least one transaction", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
		})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Only allowed 1 transaction in a resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseMultipleTx))
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)

		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
		// as there are two transaction, the Times is 2 here, possible enhancement to remove this duplicate call
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, utils.LateFilingPenaltyCompanyCode).Return(nil, nil).Times(2)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(nil).Times(2)

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
				{PenaltyRef: "A0378421", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("internal server error when failing to create payable resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		// expect the CreatePayableResource to be called once and return an error
		mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any()).Return(errors.New("any error"))

		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, utils.LateFilingPenaltyCompanyCode).Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(nil)

		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("successfully creating a payable request", t, func() {
		testCases := []struct {
			name        string
			companyCode string
			penaltyRef  string
			urlE5       string
			e5Response  string
		}{
			{
				name:        "Late Filing",
				companyCode: utils.LateFilingPenaltyCompanyCode,
				penaltyRef:  "A1234567",
				urlE5: "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" +
					utils.LateFilingPenaltyCompanyCode + "&fromDate=1990-01-01",
				e5Response: e5ResponseLateFiling,
			},
			{
				name:        "Sanctions",
				companyCode: utils.SanctionsCompanyCode,
				penaltyRef:  "P1234567",
				urlE5: "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" +
					utils.SanctionsCompanyCode + "&fromDate=1990-01-01",
				e5Response: e5ResponseSanctions,
			},
			{
				name:        "Sanctions ROE",
				companyCode: utils.SanctionsCompanyCode,
				penaltyRef:  "U1234567",
				urlE5: "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" +
					utils.SanctionsCompanyCode + "&fromDate=1990-01-01",
				e5Response: e5ResponseSanctionsRoe,
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
				mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
				// expect the CreatePayableResource to be called once and return without error
				mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any()).Return(nil)

				mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
				mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, tc.companyCode).Return(nil, nil)
				mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(nil)

				body, _ := json.Marshal(&models.PayableRequest{
					CustomerCode: "10000024",
					CreatedBy:    authentication.AuthUserDetails{},
					Transactions: []models.TransactionItem{
						{PenaltyRef: tc.penaltyRef, Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
					},
				})

				res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true)

				So(res.Code, ShouldEqual, http.StatusCreated)
				So(res.Header().Get("Content-Type"), ShouldEqual, "application/json")
			})
		}
	})
}

func setGetCompanyCodeFromTransactionMock(companyCode string) {
	mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return companyCode, nil
	}
	getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction
}
