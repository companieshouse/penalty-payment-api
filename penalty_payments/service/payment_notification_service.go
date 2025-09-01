package service

import (
	"fmt"
	"time"

	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
)

func PaymentProcessingKafkaMessage(payableResource models.PayableResource, payment *validators.PaymentInformation, context string) error {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config for penalty payments processing kafka message production: [%v]", err)
		return err
	}

	brokerAddrs := cfg.Kafka3BrokerAddr
	topic := cfg.PenaltyPaymentsProcessingTopic

	logContext := log.Data{
		"customer_code": payableResource.CustomerCode,
		"payable_ref":   payableResource.PayableRef,
		"broker_addrs":  brokerAddrs,
		"topic":         topic,
	}

	log.InfoC(context, "getting penalty payments processing kafka producer", logContext)
	kafkaProducer, err := getProducer(brokerAddrs)
	if err != nil {
		err = fmt.Errorf("error creating penalty payments processing kafka producer: [%v]", err)
		return err
	}

	penaltyPaymentsProcessingSchema, err := getSchema(cfg.SchemaRegistryURL, topic)
	if err != nil {
		err = fmt.Errorf("error getting penalty payments processing schema from schema registry: [%v]", err)
		return err
	}
	producerSchema := &avro.Schema{
		Definition: penaltyPaymentsProcessingSchema,
	}
	log.DebugC(context, "penalty payments processing avro schema", logContext, log.Data{"schema": producerSchema})

	message, err := preparePaymentProcessingKafkaMessage(*producerSchema, payableResource, payment, topic, context)
	if err != nil {
		err = fmt.Errorf("error preparing penalty payments processing kafka message with schema: [%v]", err)
		return err
	}
	log.DebugC(context, "penalty payment processing message prepared successfully", logContext, log.Data{
		"message.Value":     message.Value,
		"message.Topic":     message.Topic,
		"message.Partition": message.Partition,
		"message.Key":       message.Key,
	})

	partition, offset, err := kafkaProducer.Send(message)
	if err != nil {
		err = fmt.Errorf("failed to send penalty payments processing message: [%v]", err)
		log.Error(err, logContext)
		return err
	}
	log.InfoC(context, "successfully published penalty payments processing message", logContext, log.Data{
		"kafka_topic":     topic,
		"kafka_partition": partition,
		"kafka_offset":    offset,
	})

	return nil
}

// preparePaymentProcessingKafkaMessage generates the kafka message that is to be sent
func preparePaymentProcessingKafkaMessage(penaltyPaymentProcessingSchema avro.Schema,
	payableResource models.PayableResource, payment *validators.PaymentInformation, topic string, context string) (*producer.Message, error) {
	// Ensure payableResource contains at least one transaction
	if payableResource.Transactions == nil || len(payableResource.Transactions) == 0 {
		err := fmt.Errorf("empty transactions list in payable resource: %v", payableResource.PayableRef)
		return nil, err
	}

	companyCode, err := getCompanyCodeFromTransaction(payableResource.Transactions)
	if err != nil {
		return nil, err
	}

	penaltyPaymentProcessing := constructMessage(payableResource, companyCode, payment)

	logContext := log.Data{
		"customer_code": payableResource.CustomerCode,
		"payable_ref":   payableResource.PayableRef,
	}
	log.DebugC(context, "penalty payment processing message constructed", logContext, log.Data{
		"penalty_payments_processing": penaltyPaymentProcessing,
	})

	messageBytes, err := penaltyPaymentProcessingSchema.Marshal(penaltyPaymentProcessing)
	if err != nil {
		err = fmt.Errorf("error marshalling penalty payment processing message: [%v]", err)
		return nil, err
	}

	return &producer.Message{Value: messageBytes, Topic: topic}, nil
}

func constructMessage(payableResource models.PayableResource, companyCode string, payment *validators.PaymentInformation) models.PenaltyPaymentsProcessing {
	transactionPayments := transformToTransactionPayments(payableResource)

	penaltyPaymentProcessing := models.PenaltyPaymentsProcessing{
		Attempt:           1,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		CompanyCode:       companyCode,
		CustomerCode:      payableResource.CustomerCode,
		PaymentID:         payment.PaymentID,
		ExternalPaymentID: payment.ExternalPaymentID,
		PaymentReference:  payment.Reference,
		PaymentAmount:     payment.Amount,
		// set to the only transaction in the array until multiple payments functionality comes
		TotalValue:          transactionPayments[0].Value,
		TransactionPayments: transactionPayments,
		CardType:            payment.CardType,
		Email:               payment.CreatedBy,
		PayableRef:          payableResource.PayableRef,
	}
	return penaltyPaymentProcessing
}

func transformToTransactionPayments(payableResource models.PayableResource) []models.TransactionPayment {
	var transactionPayments []models.TransactionPayment

	for _, transaction := range payableResource.Transactions {
		transactionPayments = append(transactionPayments, models.TransactionPayment{
			TransactionReference: transaction.PenaltyRef,
			Value:                transaction.Amount,
		})
	}
	return transactionPayments
}
