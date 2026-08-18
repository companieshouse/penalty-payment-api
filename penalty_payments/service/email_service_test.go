package service

import (
	"errors"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"strconv"
	"testing"
)

var req = &http.Request{}

const AppId = "chs-monitor-notification-matcher.filing"
const MessageType = "monitor"
const MessageID = "msg-001"
const EmailAddress = "joe@bloggs.com"
const PayableReference = "late-filing-penalty-123"
const CustomerCode = "NI123456"
const PenaltyReference = "A1234567"
const Amount = 300
const PenaltyReason = "Late filing of accounts"
const CompanyName = "Test Company"

func TestSendEmailMessageViaChsKafkaApi(t *testing.T) {
	apiURL := "https://api.companieshouse.gov.uk"
	httpmock.Activate()
	// Set the prepareEmailMessage back to the actual implementation after unit tests
	t.Cleanup(func() {
		prepareEmailMessage = realPrepareEmailMessage
		httpmock.DeactivateAndReset()
	})

	Convey("Error preparing email message", t, func() {
		// Given
		prepareEmailMessage = testPrepareEmailMessageFail

		// When
		errorResult := SendEmailMessageViaChsKafkaApi(models.PayableResource{}, &http.Request{}, nil, nil, nil)

		// Then
		So(errorResult.Error(), ShouldContainSubstring, "error preparing email message")
	})

	Convey("Error sending email to api", t, func() {
		// Given
		prepareEmailMessage = testPrepareEmailMessageSuccess
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodPost, apiURL+"/send-email", httpmock.NewStringResponder(400, "{}"))

		// When
		errorResult := SendEmailMessageViaChsKafkaApi(models.PayableResource{}, &http.Request{}, nil, nil, nil)

		// Then
		So(errorResult.Error(), ShouldContainSubstring, "error sending email message")
	})

	Convey("Successfully send email message", t, func() {
		// Given
		prepareEmailMessage = testPrepareEmailMessageSuccess
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodPost, apiURL+"/send-email", httpmock.NewStringResponder(200, "{}"))

		// When
		errorResult := SendEmailMessageViaChsKafkaApi(models.PayableResource{}, &http.Request{}, nil, nil, nil)

		// Then
		So(errorResult, ShouldBeNil)
	})
}

func TestUnitRealPrepareEmailMessage(t *testing.T) {
	getConfig = testGetConfigSuccess
	getCompanyName = testGetCompanyNameSuccess
	getCompanyCodeFromTransaction = testGetCompanyCodeFromTransactionSuccess
	getPenaltyRefTypeFromTransaction = testGetPenaltyRefTypeFromTransactionSuccess
	getPayablePenalty = testGetPayablePenaltySuccess

	t.Cleanup(func() {
		getConfig = config.Get
		getCompanyName = GetCompanyName
		getCompanyCodeFromTransaction = utils.GetCompanyCodeFromTransaction
		getPenaltyRefTypeFromTransaction = utils.GetPenaltyRefTypeFromTransaction
		getPayablePenalty = api.PayablePenalty
	})

	Convey("Error getting config", t, func() {
		// Given
		getConfig = testGetConfigFail
		defer func() {
			getConfig = testGetConfigSuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(models.PayableResource{}, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error getting config")
	})

	Convey("Error getting company name", t, func() {
		// Given
		getCompanyName = testGetCompanyNameFail
		defer func() {
			getCompanyName = testGetCompanyNameSuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(models.PayableResource{}, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error getting company name")
	})

	Convey("Error when transactions list in payable resource is length 0", t, func() {
		// Given
		incorrectPayableResource := payableResourceNoItems

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(incorrectPayableResource, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "empty transactions list in payable resource")
	})

	Convey("Error getting company code from transaction", t, func() {
		// Given
		getCompanyCodeFromTransaction = testGetCompanyCodeFromTransactionFail
		defer func() {
			getCompanyCodeFromTransaction = testGetCompanyCodeFromTransactionSuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(payableResourceSuccess, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error getting company code")
	})

	Convey("Error getting penalty reference from transaction", t, func() {
		// Given
		getPenaltyRefTypeFromTransaction = testGetPenaltyRefTypeFromTransactionFail
		defer func() {
			getPenaltyRefTypeFromTransaction = testGetPenaltyRefTypeFromTransactionSuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(payableResourceSuccess, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error getting penalty ref type")
	})

	Convey("Error getting payable penalty", t, func() {
		// Given
		getPayablePenalty = testGetPayablePenaltyFail
		defer func() {
			getPayablePenalty = testGetPayablePenaltySuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(payableResourceSuccess, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error getting payable penalty")
	})

	Convey("Error parsing made up date", t, func() {
		// Given
		getPayablePenalty = testGetPayablePenaltyWrongMadeUpDate
		defer func() {
			getPayablePenalty = testGetPayablePenaltySuccess
		}()

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(payableResourceSuccess, &http.Request{}, nil, nil, nil)

		// Then
		So(preparedEmailMessage, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "error parsing made up date")
	})

	Convey("Successfully prepare email message", t, func() {
		// Given
		penaltyDetailsMap := config.PenaltyDetailsMap{
			Details: map[string]config.PenaltyDetails{
				utils.LateFilingPenaltyRefType: {
					EmailReceivedAppId: AppId,
					EmailMsgType:       MessageType,
				},
			},
		}

		// When
		preparedEmailMessage, err := realPrepareEmailMessage(payableResourceSuccess, &http.Request{}, &penaltyDetailsMap, nil, nil)

		// Then
		So(err, ShouldBeNil)
		So(preparedEmailMessage.AppID, ShouldEqual, AppId)
		So(preparedEmailMessage.MessageID, ShouldContainSubstring, PayableReference)
		So(preparedEmailMessage.MessageID, ShouldContainSubstring, "@companieshouse.gov.uk")
		So(preparedEmailMessage.MessageType, ShouldEqual, MessageType)
		So(preparedEmailMessage.Data, ShouldContainSubstring, CustomerCode)
		So(preparedEmailMessage.Data, ShouldContainSubstring, PayableReference)
		So(preparedEmailMessage.Data, ShouldContainSubstring, EmailAddress)
		So(preparedEmailMessage.Data, ShouldContainSubstring, PayableReference)
		So(preparedEmailMessage.Data, ShouldContainSubstring, "18 August 2026")
		So(preparedEmailMessage.Data, ShouldContainSubstring, strconv.Itoa(Amount))
		So(preparedEmailMessage.Data, ShouldContainSubstring, PenaltyReason)
		So(preparedEmailMessage.Data, ShouldContainSubstring, CompanyName)
		So(preparedEmailMessage.EmailAddress, ShouldEqual, EmailAddress)
		So(preparedEmailMessage.CreatedAt, ShouldNotBeNil)
	})

}

func testPrepareEmailMessageSuccess(_ models.PayableResource, _ *http.Request, _ *config.PenaltyDetailsMap, _ *models.AllowedTransactionMap, _ dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
	return &models.EmailSend{
		AppID:        AppId,
		MessageID:    MessageID,
		MessageType:  MessageType,
		Data:         "{\"customer_email\":\"joe@bloggs.com\",\"customer_feedback\":\"Great site man\",\"customer_name\":\"Joe Bloggs\"}",
		EmailAddress: EmailAddress,
	}, nil
}

func testPrepareEmailMessageFail(_ models.PayableResource, _ *http.Request, _ *config.PenaltyDetailsMap, _ *models.AllowedTransactionMap, _ dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
	return nil, errors.New("error preparing email message")
}

func testGetConfigSuccess() (*config.Config, error) {
	return &config.Config{}, nil
}

func testGetConfigFail() (*config.Config, error) {
	return nil, errors.New("error getting config")
}

func testGetCompanyNameSuccess(_ string, _ *http.Request) (string, error) {
	return CompanyName, nil
}

func testGetCompanyNameFail(_ string, _ *http.Request) (string, error) {
	return "", errors.New("error getting company name")
}

func testGetCompanyCodeFromTransactionSuccess(_ []models.TransactionItem) (string, error) {
	return utils.LateFilingPenaltyCompanyCode, nil
}

func testGetCompanyCodeFromTransactionFail(_ []models.TransactionItem) (string, error) {
	return "", errors.New("error getting company code")
}

func testGetPenaltyRefTypeFromTransactionSuccess(_ []models.TransactionItem) (string, error) {
	return utils.LateFilingPenaltyRefType, nil
}

func testGetPenaltyRefTypeFromTransactionFail(_ []models.TransactionItem) (string, error) {
	return "", errors.New("error getting penalty ref type")
}

func testGetPayablePenaltySuccess(_ types.PayablePenaltyParams) (*models.TransactionItem, error) {
	return &models.TransactionItem{
		Amount:     Amount,
		MadeUpDate: "2026-08-18",
		Reason:     PenaltyReason,
	}, nil
}

func testGetPayablePenaltyFail(_ types.PayablePenaltyParams) (*models.TransactionItem, error) {
	return nil, errors.New("error getting payable penalty")
}

func testGetPayablePenaltyWrongMadeUpDate(_ types.PayablePenaltyParams) (*models.TransactionItem, error) {
	return &models.TransactionItem{
		MadeUpDate: "incorrect",
	}, nil
}

var payableResourceNoItems = models.PayableResource{
	Transactions: []models.TransactionItem{},
}

var payableResourceSuccess = models.PayableResource{
	CreatedBy: models.CreatedBy{
		Email: EmailAddress,
	},
	CustomerCode: CustomerCode,
	PayableRef:   PayableReference,
	Transactions: []models.TransactionItem{
		{
			PenaltyRef: PenaltyReference,
		},
	},
}
