package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/filing-notification-sender/util"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

// SendEmailKafkaMessage sends a kafka message to the email-sender to send an email
func SendEmailKafkaMessage(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
	apDaoSvc dao.AccountPenaltiesDaoService) error {
	cfg, err := getConfig()
	requestId := req.Header.Get("X-Request-Id")
	if err != nil {
		err = fmt.Errorf("error getting config for kafka message production: [%v]", err)
		return err
	}

	brokerAddrs := cfg.BrokerAddr
	topic := cfg.EmailSendTopic

	logContext := log.Data{
		"customer_code": payableResource.CustomerCode,
		"payable_ref":   payableResource.PayableRef,
		"broker_addrs":  brokerAddrs,
		"topic":         topic,
	}

	log.InfoC(requestId, "getting email send kafka producer", logContext)
	kafkaProducer, err := getProducer(brokerAddrs)
	if err != nil {
		err = fmt.Errorf("error creating email send kafka producer: [%v]", err)
		return err
	}

	log.DebugC(requestId, "getting email send avro schema", logContext)
	emailSendSchema, err := getSchema(cfg.SchemaRegistryURL, topic)
	if err != nil {
		err = fmt.Errorf("error getting email send schema from schema registry: [%v]", err)
		return err
	}
	producerSchema := &avro.Schema{
		Definition: emailSendSchema,
	}
	log.DebugC(requestId, "email send avro schema", logContext, log.Data{"schema": producerSchema})

	log.InfoC(requestId, "preparing email send message", logContext)
	message, err := prepareEmailKafkaMessage(
		*producerSchema, payableResource, req, penaltyDetailsMap, apDaoSvc, topic)
	if err != nil {
		err = fmt.Errorf("error preparing email send kafka message with schema: [%v]", err)
		return err
	}

	log.DebugC(requestId, "email send message prepared successfully", log.Data{
		"message.Value":     message.Value,
		"message.Topic":     message.Topic,
		"message.Partition": message.Partition,
		"message.Key":       message.Key,
	})

	partition, offset, err := kafkaProducer.Send(message)
	if err != nil {
		err = fmt.Errorf("failed to send email send message: [%v]", err)
		log.ErrorC(requestId, err, logContext)
		return err
	}
	log.InfoC(requestId, "successfully published email send message", logContext, log.Data{
		"kafka_topic":     topic,
		"kafka_partition": partition,
		"kafka_offset":    offset,
	})

	return nil
}

// prepareEmailKafkaMessage generates the kafka message that is to be sent
func prepareEmailKafkaMessage(emailSendSchema avro.Schema, payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
	apDaoSvc dao.AccountPenaltiesDaoService, topic string) (*producer.Message, error) {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config: [%v]", err)
		return nil, err
	}

	companyName, err := getCompanyName(payableResource.CustomerCode, req)
	if err != nil {
		err = fmt.Errorf("error getting company name: [%v]", err)
		return nil, err
	}

	// Ensure payableResource contains at least one transaction
	if payableResource.Transactions == nil || len(payableResource.Transactions) == 0 {
		err = fmt.Errorf("empty transactions list in payable resource: %v", payableResource.PayableRef)
		return nil, err
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
		AccountPenaltiesDaoService: apDaoSvc,
		RequestId:                  "",
	}
	payablePenalty, err := getPayablePenalty(params)
	if err != nil {
		err = fmt.Errorf("error getting transaction for penalty: [%v]", err)
		return nil, err
	}

	// Convert madeUpDate to readable format for email
	madeUpDate, err := time.Parse("2006-01-02", payablePenalty.MadeUpDate)
	if err != nil {
		err = fmt.Errorf("error parsing made up date: [%v]", err)
		return nil, err
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
		Subject:           fmt.Sprintf("Confirmation of your Companies House penalty payment"),
		CHSURL:            cfg.CHSURL,
	}

	requestId := req.Header.Get("X-Request-Id")

	logContext := log.Data{
		"customer_code": payableResource.CustomerCode,
		"payable_ref":   payableResource.PayableRef,
	}
	log.DebugC(requestId, "email send message data field", logContext, log.Data{"data_field": dataFieldMessage})

	dataBytes, err := json.Marshal(dataFieldMessage)
	if err != nil {
		err = fmt.Errorf("error marshalling dataFieldMessage: [%v]", err)
		return nil, err
	}

	messageID := "<" + payableResource.PayableRef + "." + strconv.Itoa(util.Random(0, 100000)) + "@companieshouse.gov.uk>"

	emailSendMessage := models.EmailSend{
		AppID:        penaltyDetailsMap.Details[penaltyRefType].EmailReceivedAppId,
		MessageID:    messageID,
		MessageType:  penaltyDetailsMap.Details[penaltyRefType].EmailMsgType,
		Data:         string(dataBytes),
		EmailAddress: payableResource.CreatedBy.Email,
		CreatedAt:    time.Now().String(),
	}

	log.DebugC(requestId, "email send message", logContext, log.Data{"email_send": emailSendMessage})

	messageBytes, err := emailSendSchema.Marshal(emailSendMessage)
	if err != nil {
		err = fmt.Errorf("error marshalling email send message: [%v]", err)
		return nil, err
	}

	return &producer.Message{Value: messageBytes, Topic: topic}, nil
}
