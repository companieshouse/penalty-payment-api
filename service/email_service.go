package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/filing-notification-sender/util"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/utils"
)

// ProducerTopic is the topic to which the email-send kafka message is sent
const ProducerTopic = "email-send"

// ProducerSchemaName is the schema which will be used to send the email-send kafka message with
const ProducerSchemaName = "email-send"

var getConfig = func() (*config.Config, error) {
	return config.Get()
}
var getProducer = func(config *config.Config) (*producer.Producer, error) {
	return producer.New(&producer.Config{Acks: &producer.WaitForAll, BrokerAddrs: config.BrokerAddr})
}
var getSchema = func(url string) (string, error) {
	return schema.Get(url, ProducerSchemaName)
}

// SendEmailKafkaMessage sends a kafka message to the email-sender to send an email
func SendEmailKafkaMessage(payableResource models.PayableResource, req *http.Request,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) error {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config for kafka message production: [%v]", err)
		return err
	}

	// Get a producer
	kafkaProducer, err := getProducer(cfg)
	if err != nil {
		err = fmt.Errorf("error creating kafka producer: [%v]", err)
		return err
	}
	emailSendSchema, err := getSchema(cfg.SchemaRegistryURL)
	if err != nil {
		err = fmt.Errorf("error getting schema from schema registry: [%v]", err)
		return err
	}
	producerSchema := &avro.Schema{
		Definition: emailSendSchema,
	}

	// Prepare a message with the avro schema
	message, err := prepareKafkaMessage(*producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error preparing kafka message with schema: [%v]", err)
		return err
	}

	// Send the message
	partition, offset, err := kafkaProducer.Send(message)
	if err != nil {
		err = fmt.Errorf("failed to send message in partition: %d at offset %d", partition, offset)
		return err
	}
	return nil
}

var getCompanyCodeFromTransaction = func(transactions []models.TransactionItem) (string, error) {
	return utils.GetCompanyCodeFromTransaction(transactions)
}

var getCompanyName = func(companyNumber string, req *http.Request) (string, error) {
	return GetCompanyName(companyNumber, req)
}

var getTransactionForPenalty = func(companyNumber, companyCode, penaltyReference string, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListItem, error) {
	return GetTransactionForPenalty(companyNumber, companyCode, penaltyReference, penaltyDetailsMap, allowedTransactionsMap)
}

// prepareKafkaMessage generates the kafka message that is to be sent
func prepareKafkaMessage(emailSendSchema avro.Schema, payableResource models.PayableResource, req *http.Request,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*producer.Message, error) {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config: [%v]", err)
		return nil, err
	}

	// Access Company Name to be included in the email
	companyName, err := getCompanyName(payableResource.CompanyNumber, req)
	if err != nil {
		err = fmt.Errorf("error getting company name: [%v]", err)
		return nil, err
	}

	companyCode, err := getCompanyCodeFromTransaction(payableResource.Transactions)
	if err != nil {
		return nil, err
	}

	// Access specific transaction that was paid for
	payedTransaction, err := getTransactionForPenalty(payableResource.CompanyNumber, companyCode,
		payableResource.Transactions[0].TransactionID, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error getting transaction for penalty: [%v]", err)
		return nil, err
	}

	// Convert madeUpDate to readable format for email
	madeUpDate, err := time.Parse("2006-01-02", payedTransaction.MadeUpDate)
	if err != nil {
		err = fmt.Errorf("error parsing made up date: [%v]", err)
		return nil, err
	}

	// Set dataField to be used in the avro schema.
	dataFieldMessage := models.DataField{
		PayableResource:   payableResource,
		TransactionID:     payableResource.Transactions[0].TransactionID,
		MadeUpDate:        madeUpDate.Format("2 January 2006"),
		TransactionDate:   time.Now().Format("2 January 2006"),
		Amount:            fmt.Sprintf("%g", payedTransaction.OriginalAmount),
		CompanyName:       companyName,
		FilingDescription: payedTransaction.Reason,
		To:                payableResource.CreatedBy.Email,
		Subject:           fmt.Sprintf("Confirmation of your Companies House penalty payment"),
		CHSURL:            cfg.CHSURL,
	}

	dataBytes, err := json.Marshal(dataFieldMessage)
	if err != nil {
		err = fmt.Errorf("error marshalling dataFieldMessage: [%v]", err)
		return nil, err
	}

	messageID := "<" + payableResource.Reference + "." + strconv.Itoa(util.Random(0, 100000)) + "@companieshouse.gov.uk>"

	emailSendMessage := models.EmailSend{
		AppID:        penaltyDetailsMap.Details[companyCode].EmailReceivedAppId,
		MessageID:    messageID,
		MessageType:  penaltyDetailsMap.Details[companyCode].EmailMsgType,
		Data:         string(dataBytes),
		EmailAddress: payableResource.CreatedBy.Email,
		CreatedAt:    time.Now().String(),
	}

	messageBytes, err := emailSendSchema.Marshal(emailSendMessage)
	if err != nil {
		err = fmt.Errorf("error marshalling email send message: [%v]", err)
		return nil, err
	}

	producerMessage := &producer.Message{
		Value: messageBytes,
		Topic: ProducerTopic,
	}
	return producerMessage, nil
}
