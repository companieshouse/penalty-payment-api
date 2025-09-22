package e5

import (
	"net/http"
	"testing"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func hasFieldError(field, tag string, errs validator.ValidationErrors) bool {
	for _, e := range errs {
		f := e.Field()
		t := e.Tag()
		if f == field && t == tag {
			return true
		}
	}
	return false
}

func getE5Client() ClientInterface {
	return NewClient("foo", "https://e5")
}

var requestId = "123456abc"

type TestCase struct {
	name       string
	statusCode int
	payload    string
	err        error
}

func TestUnitClient_CreatePayment(t *testing.T) {
	e5 := getE5Client()
	url := "https://e5/arTransactions/payment?ADV_userName=foo"

	Convey("creating a payment", t, func() {
		input := &CreatePaymentInput{
			CompanyCode:  utils.LateFilingPenaltyCompanyCode,
			CustomerCode: "1000024",
			PaymentID:    "1234",
			TotalValue:   100,
			Transactions: []*CreatePaymentTransaction{
				{
					TransactionReference: "1234",
					Value:                100,
				},
			},
		}

		testCases := []TestCase{
			{
				name:       "response should be unsuccessful when there is a 500 error from E5",
				statusCode: http.StatusInternalServerError,
				payload:    "test error",
				err:        ErrE5InternalServer,
			},
			{
				name:       "response should be unsuccessful when the company does not exist",
				statusCode: http.StatusNotFound,
				payload:    "company not found",
				err:        ErrE5NotFound,
			},
			{
				name:       "response should be successful if a 200 is returned from E5",
				statusCode: http.StatusOK,
				payload:    "",
				err:        nil,
			},
		}

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		for _, testCase := range testCases {
			Convey(testCase.name, func() {
				if testCase.statusCode == http.StatusOK {
					responder, _ := httpmock.NewJsonResponder(testCase.statusCode, nil)
					httpmock.RegisterResponder(http.MethodPost, url, responder)

					err := e5.CreatePayment(input, requestId)

					So(err, ShouldBeNil)
				} else {
					httpErr := &apiErrorResponse{Code: testCase.statusCode, Message: testCase.payload}
					responder, _ := httpmock.NewJsonResponder(testCase.statusCode, httpErr)
					httpmock.RegisterResponder(http.MethodPost, url, responder)

					err := e5.CreatePayment(input, requestId)

					So(err, ShouldBeError, testCase.err)
				}
			})
		}
	})
}

// this is the response returned by e5 when the company number is incorrect i.e. no transactions exist
var e5EmptyResponse = `
{
  "page" : {
    "size" : 0,
    "totalElements" : 0,
    "totalPages" : 1,
    "number" : 0
  },
  "data" : [ ]
}`

var e5TransactionResponse = `
{
  "page" : {
    "size" : 1,
    "totalElements" : 1,
    "totalPages" : 1,
    "number" : 0
  },
  "data" : [ {
    "companyCode" : "LP",
    "ledgerCode" : "EW",
    "customerCode" : "10000024",
    "transactionReference" : "00378420",
    "transactionDate" : "2017-11-28",
    "madeUpDate" : "2017-02-28",
    "amount" : 150,
    "outstandingAmount" : 150,
    "isPaid" : false,
    "transactionType" : "1",
    "transactionSubType" : "EU",
    "typeDescription" : "Penalty Ltd Wel & Eng <=1m     LTDWA    ",
    "dueDate" : "2017-12-12",
    "accountStatus": "CHS",
    "dunningStatus": "PEN1"
  }]
}
`

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

func getTestE5Transactions(response string, statusCode int) (*GetTransactionsResponse, error) {
	e5 := getE5Client()
	url := "https://e5/arTransactions/10000024?ADV_userName=foo&companyCode=LP&fromDate=1990-01-01"
	transactionInput := &GetTransactionsInput{CustomerCode: "10000024", CompanyCode: "LP"}

	responder := httpmock.NewStringResponder(statusCode, response)
	httpmock.RegisterResponder(http.MethodGet, url, responder)

	return e5.GetTransactions(transactionInput, requestId)
}

func TestUnitClient_GetTransactions(t *testing.T) {
	Convey("getting a list of transactions for a company", t, func() {

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("company does not exist or no transactions returned", func() {
			r, err := getTestE5Transactions(e5EmptyResponse, http.StatusOK)

			So(err, ShouldBeNil)
			So(r.Transactions, ShouldBeEmpty)
			So(r.Page.Size, ShouldEqual, 0)
		})

		Convey("should return a list of transactions", func() {
			r, err := getTestE5Transactions(e5TransactionResponse, http.StatusOK)

			So(err, ShouldBeNil)
			So(r.Transactions, ShouldHaveLength, 1)
		})

		Convey("using an incorrect company code", func() {
			r, err := getTestE5Transactions(e5ValidationError, http.StatusBadRequest)

			So(r, ShouldBeNil)
			So(err, ShouldBeError, ErrE5BadRequest)
		})

	})
}

func getAuthoriseConfirmTestCases() []TestCase {
	return []TestCase{
		{
			name:       "500 error from E5",
			statusCode: http.StatusInternalServerError,
			payload:    e5ValidationError,
			err:        ErrE5InternalServer,
		},
		{
			name:       "400 error from E5",
			statusCode: http.StatusBadRequest,
			payload:    e5ValidationError,
			err:        ErrE5BadRequest,
		},
		{
			name:       "404 error from E5",
			statusCode: http.StatusNotFound,
			payload:    e5ValidationError,
			err:        ErrE5NotFound,
		},
		{
			name:       "403 error from E5",
			statusCode: http.StatusForbidden,
			payload:    e5ValidationError,
			err:        ErrUnexpectedServerError,
		},
		{
			name:       "successful request",
			statusCode: http.StatusOK,
			payload:    "",
			err:        nil,
		},
	}
}

func TestUnitClient_AuthorisePayment(t *testing.T) {
	e5 := getE5Client()

	Convey("email, paymentId are required parameters", t, func() {
		input := &AuthorisePaymentInput{}

		err := e5.AuthorisePayment(input, requestId)

		So(err, ShouldNotBeNil)

		errors := err.(validator.ValidationErrors)

		So(errors, ShouldHaveLength, 3)
		So(hasFieldError("Email", "required", errors), ShouldBeTrue)
		So(hasFieldError("PaymentID", "required", errors), ShouldBeTrue)
		So(hasFieldError("CompanyCode", "required", errors), ShouldBeTrue)
	})

	url := "https://e5/arTransactions/payment/authorise?ADV_userName=foo"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, testCase := range getAuthoriseConfirmTestCases() {
		Convey(testCase.name, t, func() {

			responder := httpmock.NewStringResponder(testCase.statusCode, testCase.payload)
			httpmock.RegisterResponder(http.MethodPost, url, responder)

			err := e5.AuthorisePayment(&AuthorisePaymentInput{PaymentID: "123", Email: "test@example.com", CompanyCode: "LP"}, requestId)

			if testCase.err == nil {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.err)
			}
		})
	}
}

func TestUnitClient_Confirm(t *testing.T) {
	e5 := getE5Client()

	Convey("paymentId is required", t, func() {
		err := e5.ConfirmPayment(&PaymentActionInput{}, "")

		errors := err.(validator.ValidationErrors)

		So(err, ShouldNotBeNil)
		So(hasFieldError("PaymentID", "required", errors), ShouldBeTrue)
	})

	url := "https://e5/arTransactions/payment/confirm?ADV_userName=foo"
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, testCase := range getAuthoriseConfirmTestCases() {
		Convey(testCase.name, t, func() {

			responder := httpmock.NewStringResponder(testCase.statusCode, testCase.payload)
			httpmock.RegisterResponder(http.MethodPost, url, responder)

			err := e5.ConfirmPayment(&PaymentActionInput{PaymentID: "123", CompanyCode: "LP"}, requestId)

			if testCase.err == nil {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldBeError, testCase.err)
			}
		})
	}
}
