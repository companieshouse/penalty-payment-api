package service

import (
	"fmt"

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

	kafkaProducer, err := getProducer(cfg)
	if err != nil {
		err = fmt.Errorf("error creating kafka producer: [%v]", err)
		return err
	}
	paymentProcessingSendSchema, err := getSchema(cfg.SchemaRegistryURL, "penalty-payments-processing")
	if err != nil {
		err = fmt.Errorf("error getting schema from schema registry: [%v]", err)
		return err
	}
	producerSchema := &avro.Schema{
		Definition: paymentProcessingSendSchema,
	}

	message, err := preparePaymentProcessingKafkaMessage(*producerSchema, payableResource, payment)
	if err != nil {
		err = fmt.Errorf("error preparing kafka message with schema: [%v]", err)
		return err
	}

	partition, offset, err := kafkaProducer.Send(message)
	if err != nil {
		err = fmt.Errorf("failed to send message in partition: %d at offset %d", partition, offset)
		return err
	}
	return nil
}

// prepareEmailKafkaMessage generates the kafka message that is to be sent
func preparePaymentProcessingKafkaMessage(penaltyPaymentProcessingSchema avro.Schema,
	payableResource models.PayableResource, payment *validators.PaymentInformation) (*producer.Message, error) {
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

	log.Info("construct message", log.Data{
		"Attempt":           penaltyPaymentProcessing.Attempt,
		"CompanyCode":       penaltyPaymentProcessing.CompanyCode,
		"CustomerCode":      penaltyPaymentProcessing.CustomerCode,
		"PaymentId":         penaltyPaymentProcessing.PaymentId,
		"PaymentExternalId": penaltyPaymentProcessing.PaymentExternalId,
		"PaymentReference":  penaltyPaymentProcessing.PaymentReference,
		"PaymentAmount":     penaltyPaymentProcessing.PaymentAmount,
		"TransactionPayments - TransactionReference": penaltyPaymentProcessing.TransactionPayments[0].TransactionReference,
		"TransactionPayments - Value":                penaltyPaymentProcessing.TransactionPayments[0].Value,
		"CardType":                                   penaltyPaymentProcessing.CardType,
		"Email":                                      penaltyPaymentProcessing.Email,
		"PayableRef":                                 penaltyPaymentProcessing.PayableRef,
	})

	messageBytes, err := penaltyPaymentProcessingSchema.Marshal(penaltyPaymentProcessing)
	if err != nil {
		err = fmt.Errorf("error marshalling penalty payment processing message: [%v]", err)
		return nil, err
	}

	return &producer.Message{Value: messageBytes, Topic: "penalty-payments-processing"}, nil
}

func constructMessage(payableResource models.PayableResource, companyCode string, payment *validators.PaymentInformation) PenaltyPaymentProcessingSend {
	transactionPayments := transformToTransactionPayments(payableResource)

	penaltyPaymentProcessing := PenaltyPaymentProcessingSend{
		Attempt:             1,
		CompanyCode:         companyCode,
		CustomerCode:        payableResource.CustomerCode,
		PaymentId:           payment.PaymentID,
		PaymentExternalId:   payment.ExternalPaymentID,
		PaymentReference:    payment.Reference,
		PaymentAmount:       payment.Amount,
		TransactionPayments: transactionPayments,
		CardReference:       payment.Reference,
		CardType:            payment.CardType,
		Email:               payment.CreatedBy,
		PayableRef:          payableResource.PayableRef,
	}
	return penaltyPaymentProcessing
}

func transformToTransactionPayments(payableResource models.PayableResource) []TransactionPayment {
	var transactionPayments []TransactionPayment

	for _, transaction := range payableResource.Transactions {
		transactionPayments = append(transactionPayments, TransactionPayment{
			transaction.PenaltyRef,
			transaction.Amount,
		})
	}
	return transactionPayments
}

type PenaltyPaymentProcessingSend struct {
	Attempt             int                  `avro:"attempt"`
	CompanyCode         string               `avro:"company_code"`
	CustomerCode        string               `avro:"customer_code"`
	PaymentId           string               `avro:"payment_id"`
	PaymentExternalId   string               `avro:"payment_external_id"`
	PaymentReference    string               `avro:"payment_reference"`
	PaymentAmount       string               `avro:"payment_amount"`
	TotalValue          float64              `avro:"total_value"`
	TransactionPayments []TransactionPayment `avro:"transaction_payments"`
	CardReference       string               `avro:"card_reference"`
	CardType            string               `avro:"card_type"`
	Email               string               `avro:"email"`
	PayableRef          string               `avro:"payable_ref"`
}

type TransactionPayment struct {
	TransactionReference string  `avro:"transaction_reference"`
	Value                float64 `avro:"value"`
}
