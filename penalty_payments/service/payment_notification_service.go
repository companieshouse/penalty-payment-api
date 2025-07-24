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

func PaymentProcessingKafkaMessage(payableResource models.PayableResource, payment *validators.PaymentInformation) error {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config for kafka message production: [%v]", err)
		return err
	}

	log.Debug(fmt.Sprintf("config data: %+v", cfg))
	topic := cfg.PenaltyPaymentsProcessingTopic

	log.Info("getting penalty payments processing kafka producer", log.Data{"customer_code": payableResource.CustomerCode})
	kafkaProducer, err := getProducer(cfg)
	if err != nil {
		err = fmt.Errorf("error creating kafka producer: [%v]", err)
		return err
	}

	paymentProcessingSendSchema, err := getSchema(cfg.SchemaRegistryURL, topic)
	if err != nil {
		err = fmt.Errorf("error getting schema from schema registry: [%v]", err)
		return err
	}
	producerSchema := &avro.Schema{
		Definition: paymentProcessingSendSchema,
	}

	message, err := preparePaymentProcessingKafkaMessage(*producerSchema, payableResource, payment, topic)
	if err != nil {
		err = fmt.Errorf("error preparing kafka message with schema: [%v]", err)
		return err
	}
	log.Info("payment processing message prepared successfully", log.Data{
		"message.Value":     message.Value,
		"message.Topic":     message.Topic,
		"message.Partition": message.Partition,
		"message.Key":       message.Key,
	})

	partition, offset, err := kafkaProducer.Send(message)
	if err != nil {
		err = fmt.Errorf("failed to send message in partition: %d at offset %d", partition, offset)
		return err
	}

	return nil
}

// preparePaymentProcessingKafkaMessage generates the kafka message that is to be sent
func preparePaymentProcessingKafkaMessage(penaltyPaymentProcessingSchema avro.Schema,
	payableResource models.PayableResource, payment *validators.PaymentInformation, topic string) (*producer.Message, error) {
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

	log.Debug("penalty payment processing message", log.Data{
		"attempt":             penaltyPaymentProcessing.Attempt,
		"created_at":          penaltyPaymentProcessing.CreatedAt,
		"company_code":        penaltyPaymentProcessing.CompanyCode,
		"customer_code":       penaltyPaymentProcessing.CustomerCode,
		"payment_id":          penaltyPaymentProcessing.PaymentID,
		"external_payment_id": penaltyPaymentProcessing.ExternalPaymentID,
		"payment_reference":   penaltyPaymentProcessing.PaymentReference,
		"payment_amount":      penaltyPaymentProcessing.PaymentAmount,
		"transactionPayments[0] - transaction_reference": penaltyPaymentProcessing.TransactionPayments[0].TransactionReference,
		"transaction_payments[0] - value":                penaltyPaymentProcessing.TransactionPayments[0].Value,
		"card_type":                                      penaltyPaymentProcessing.CardType,
		"email":                                          penaltyPaymentProcessing.Email,
		"payable_ref":                                    penaltyPaymentProcessing.PayableRef,
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
