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
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/configctx"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func createPayableResourceHandlerTestSetup(t *testing.T) (*gomock.Controller, *mocks.MockPayableResourceDaoService, *mocks.MockAccountPenaltiesDaoService, string, error) {
	err := os.Chdir("..")
	if err != nil {
		return nil, nil, nil, "", err
	}
	cfg, _ := config.Get()
	cfg.E5APIURL = "https://e5"
	cfg.E5Username = "SYSTEM"
	url := "https://e5/arTransactions/10000024?ADV_userName=SYSTEM&companyCode=" + utils.LateFilingPenaltyCompanyCode + "&fromDate=1990-01-01"

	httpmock.Activate()
	mockCtrl := gomock.NewController(t)
	mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
	mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)

	return mockCtrl, mockPrDaoSvc, mockApDaoSvc, url, nil
}

func serveCreatePayableResourceHandler(body []byte, payableResourceService dao.PayableResourceDaoService,
	apDaoSvc dao.AccountPenaltiesDaoService, withAuthUserDetails bool, customerCode string) *httptest.ResponseRecorder {
	template := "/company/%s/penalties/payable"
	path := fmt.Sprintf(template, customerCode)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	res := httptest.NewRecorder()

	baseCtx := testContext(withAuthUserDetails, customerCode)

	penaltyDetailsMap := &config.PenaltyDetailsMap{
		Name: "penalty details",
		Details: map[string]config.PenaltyDetails{
			utils.LateFilingPenaltyRefType: {
				Description:        "Late Filing Penalty",
				DescriptionId:      "late-filing-penalty",
				ClassOfPayment:     "penalty-lfp",
				ResourceKind:       "late-filing-penalty#late-filing-penalty",
				ProductType:        "late-filing-penalty",
				EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
				EmailMsgType:       "penalty_payment_received_email",
			},
		},
	}

	allowedTransactionMap := &models.AllowedTransactionMap{
		Types: map[string]map[string]bool{
			"1": {
				"EJ": true,
				"EK": true,
				"EL": true,
				"EU": true,
				"S1": true,
				"A2": true,
			},
		},
	}

	// Wrap base context with ConfigContext
	ctxWithConfig := configctx.WithConfig(
		baseCtx,
		[]finance_config.FinancePenaltyTypeConfig{},
		[]finance_config.FinancePayablePenaltyConfig{},
		penaltyDetailsMap,
		allowedTransactionMap,
	)

	// Attach combined context to request
	req = req.WithContext(ctxWithConfig)

	handler := CreatePayableResourceHandler(payableResourceService, apDaoSvc)
	handler.ServeHTTP(res, req)

	return res
}

func testContext(withAuthUserDetails bool, customerCode string) context.Context {
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
var customerCode = "10000024"
var penaltyRef1 = "A1234567"
var penaltyRef2 = "A0378421"

func buildRequestBody(customerCode string, malformed bool, empty bool, penaltyRefs []string) []byte {
	if malformed {
		return []byte{'{'}
	} else if empty {
		body, _ := json.Marshal(&models.PayableRequest{})
		return body
	}
	payableRequest := models.PayableRequest{
		CustomerCode: customerCode,
		CreatedBy:    authentication.AuthUserDetails{},
	}
	if len(penaltyRefs) > 0 {
		var transactions []models.TransactionItem
		for _, ref := range penaltyRefs {
			transactions = append(transactions, models.TransactionItem{
				PenaltyRef: ref, Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty",
			})
		}
		payableRequest.Transactions = transactions
	}
	body, _ := json.Marshal(payableRequest)

	return body
}

func TestUnitCreatePayableResourceHandler(t *testing.T) {
	mockCtrl, mockPrDaoSvc, mockApDaoSvc, url, err := createPayableResourceHandlerTestSetup(t)
	if err != nil {
		return
	}

	defer httpmock.DeactivateAndReset()
	defer mockCtrl.Finish()

	Convey("Error decoding request body", t, func() {
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := buildRequestBody("", true, false, []string{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true, customerCode)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when user details not in context", t, func() {
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := buildRequestBody("", false, true, []string{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), false, customerCode)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when customer code not in context", t, func() {
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := buildRequestBody("", false, true, []string{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true, "")

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company code cannot be resolved", t, func() {
		getCompanyCodeFromTransaction = func(transactions []models.TransactionItem) (string, error) {
			return "", errors.New("no penalty reference found")
		}

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := buildRequestBody("", false, true, []string{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true, "")

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need at least one transaction", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		body := buildRequestBody(customerCode, false, false, []string{})
		res := serveCreatePayableResourceHandler(body, mocks.NewMockPayableResourceDaoService(mockCtrl),
			mocks.NewMockAccountPenaltiesDaoService(mockCtrl), true, customerCode)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Only allowed 1 transaction in a resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseMultipleTx))

		// as there are two transaction, the Times is 2 here, possible enhancement to remove this duplicate call
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, utils.LateFilingPenaltyCompanyCode, "").Return(nil, nil).Times(2)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil).Times(2)

		body := buildRequestBody(customerCode, false, false, []string{penaltyRef1, penaltyRef2})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true, customerCode)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("internal server error when failing to create payable resource", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseLateFiling))

		mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any(), "").Return(errors.New("any error"))
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, utils.LateFilingPenaltyCompanyCode, "").Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil)

		body := buildRequestBody(customerCode, false, false, []string{penaltyRef1})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true, customerCode)

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

				httpmock.RegisterResponder("GET", tc.urlE5, httpmock.NewStringResponder(200, tc.e5Response))

				mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any(), "").Return(nil)
				mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, tc.companyCode, "").Return(nil, nil)
				mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil)

				body := buildRequestBody(customerCode, false, false, []string{tc.penaltyRef})

				res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true, customerCode)

				So(res.Code, ShouldEqual, http.StatusCreated)
				So(res.Header().Get("Content-Type"), ShouldEqual, "application/json")
			})
		}
	})
}

func TestUnitCreatePayableResourceHandler_MockedPayablePenalty(t *testing.T) {
	mockCtrl, mockPrDaoSvc, mockApDaoSvc, url, err := createPayableResourceHandlerTestSetup(t)
	if err != nil {
		return
	}

	defer httpmock.DeactivateAndReset()
	defer mockCtrl.Finish()

	Convey("Error getting account penalties", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenaltyCompanyCode)

		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, e5ResponseMultipleTx))

		payablePenalty = func(params types.PayablePenaltyParams, penaltyConfig configctx.ConfigContext) (*models.TransactionItem, error) {
			return nil, errors.New("error")
		}

		body := buildRequestBody(customerCode, false, false, []string{penaltyRef1, penaltyRef2})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, mockApDaoSvc, true, customerCode)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func setGetCompanyCodeFromTransactionMock(companyCode string) {
	mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return companyCode, nil
	}
	getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction
}
