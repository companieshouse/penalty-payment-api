package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/filing-notification-sender/util"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var prepareEmailMessage = realPrepareEmailMessage
var newRequestFunc = http.NewRequest
var httpClient = &http.Client{}

func SendEmailMessageViaChsKafkaApi(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) error {

	requestId := log.Context(req)

	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("error getting config for sending email message chs-kafka-api: [%v]", err)
	}

	message, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, apDaoSvc)
	if err != nil || message == nil {
		return fmt.Errorf("error preparing email message for chs-kafka-api: [%v]", err)
	}

	baseURL, err := url.Parse(fmt.Sprintf("%s/send-email", cfg.ChsKafkaApiURL))
	if err != nil {
		return fmt.Errorf("invalid base URL: [%v]", err)
	}

	log.DebugC(requestId, "email send message prepared successfully", log.Data{"message": message})

	params := url.Values{}
	params.Set("app_id", message.AppID)
	params.Set("message_id", message.MessageID)
	params.Set("message_type", message.MessageType)
	params.Set("json_data", message.Data)
	params.Set("email_address", message.EmailAddress)

	baseURL.RawQuery = params.Encode()

	request, err := newRequestFunc(http.MethodPost, baseURL.String(), nil)

	if err != nil {
		return fmt.Errorf("error creating POST request to chs-kafka-api: [%v]", err)
	}

	request.Header.Add("Authorization", cfg.ChsKafkaApiKey)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpClient.Do(request)

	if err != nil && (response == nil || response.StatusCode != http.StatusAccepted) {
		logContext := log.Data{
			"customer_code": payableResource.CustomerCode,
			"payable_ref":   payableResource.PayableRef,
		}
		err = fmt.Errorf("failed to send email send message: [%v]", err)
		log.ErrorC(requestId, err, logContext)
		return err
	}

	log.InfoC(requestId, "Successfully sent email message to chs-kafka-api")

	return nil
}

func realPrepareEmailMessage(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting config: [%v]", err)
	}

	companyName, err := getCompanyName(payableResource.CustomerCode, req)
	if err != nil {
		return nil, fmt.Errorf("error getting company name: [%v]", err)
	}

	if len(payableResource.Transactions) == 0 {
		return nil, fmt.Errorf("empty transactions list in payable resource: %v", payableResource.PayableRef)
	}

	companyCode, err := getCompanyCodeFromTransaction(payableResource.Transactions)
	if err != nil {
		return nil, err
	}

	penaltyRefType, err := getPenaltyRefTypeFromTransaction(payableResource.Transactions)
	if err != nil {
		return nil, err
	}

	transaction := payableResource.Transactions[0]
	params := types.PayablePenaltyParams{
		PenaltyRefType:             penaltyRefType,
		CustomerCode:               payableResource.CustomerCode,
		CompanyCode:                companyCode,
		Transaction:                transaction,
		PenaltyDetailsMap:          penaltyDetailsMap,
		AllowedTransactionsMap:     allowedTransactionsMap,
		AccountPenaltiesDaoService: apDaoSvc,
		RequestId:                  "",
	}
	payablePenalty, err := getPayablePenalty(params)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction for penalty: [%v]", err)
	}

	madeUpDate, err := time.Parse("2006-01-02", payablePenalty.MadeUpDate)
	if err != nil {
		return nil, fmt.Errorf("error parsing made up date: [%v]", err)
	}

	dataFieldMessage := models.DataField{
		PayableResource:   payableResource,
		PenaltyRef:        payableResource.Transactions[0].PenaltyRef,
		MadeUpDate:        madeUpDate.Format("2 January 2006"),
		TransactionDate:   time.Now().Format("2 January 2006"),
		Amount:            fmt.Sprintf("%g", payablePenalty.Amount),
		CompanyName:       companyName,
		FilingDescription: payablePenalty.Reason,
		To:                payableResource.CreatedBy.Email,
		Subject:           "Confirmation of your Companies House penalty payment",
		CHSURL:            cfg.CHSURL,
	}
	requestId := log.Context(req)

	logContext := log.Data{
		"customer_code": payableResource.CustomerCode,
		"payable_ref":   payableResource.PayableRef,
	}

	log.DebugC(requestId, "email send message data field", logContext, log.Data{"data_field": dataFieldMessage})

	jsonData, err := json.Marshal(dataFieldMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshalling dataFieldMessage: [%v]", err)
	}

	messageID := "<" + payableResource.PayableRef + "." + strconv.Itoa(util.Random(0, 100000)) + "@companieshouse.gov.uk>"

	return &models.EmailSend{
		AppID:        penaltyDetailsMap.Details[penaltyRefType].EmailReceivedAppId,
		MessageID:    messageID,
		MessageType:  penaltyDetailsMap.Details[penaltyRefType].EmailMsgType,
		Data:         string(jsonData),
		EmailAddress: payableResource.CreatedBy.Email,
		CreatedAt:    time.Now().String(),
	}, nil
}
