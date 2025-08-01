package consumer

import (
	"errors"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

var (
	penaltyPayment = models.PenaltyPaymentsProcessing{
		Attempt:           1,
		CompanyCode:       "C1",
		CustomerCode:      "OE123456",
		PaymentID:         "KIYLUq1pRVuiLNA",
		ExternalPaymentID: "a8n3vp4uo1o7mf7pp2mtab7ne9",
		PaymentReference:  "financial_penalty_SQ33133143",
		PaymentAmount:     "350.00",
		TotalValue:        350.0,
		TransactionPayments: []models.TransactionPayment{{
			TransactionReference: "U1234567",
			Value:                350.0,
		}},
		CardType:   "Visa",
		Email:      "test@example.com",
		PayableRef: "SQ33133143",
	}
	e5PaymentID = "XKIYLUq1pRVuiLNA"
	cfg         = &config.Config{}
)

type mockPenaltyFinancePayment struct {
	mock.Mock
}

func (m *mockPenaltyFinancePayment) ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing,
	e5PaymentID string, cfg *config.Config) error {
	args := m.Called(penaltyPayment, e5PaymentID, cfg)
	return args.Error(0)
}

func TestUnitHandleMessage_Success(t *testing.T) {
	Convey("Handle message penalty payments processing Success", t, func() {
		// Given
		avroSchema := getAvroSchema()
		message := getConsumerMessage(avroSchema, penaltyPayment)
		mockFinancePayment := new(mockPenaltyFinancePayment)
		mockFinancePayment.On("ProcessFinancialPenaltyPayment", penaltyPayment, e5PaymentID, cfg).Return(nil)

		// When
		err := handleMessage(avroSchema, message, mockFinancePayment, cfg)

		// Then
		So(err, ShouldBeNil)
		mockFinancePayment.AssertExpectations(t)
	})
}

func TestUnitHandleMessage_UnmarshalFails(t *testing.T) {
	Convey("Handle message penalty payments processing Unmarshal fails", t, func() {
		// Given
		kafkaSchema := `{"type":"record","name":"PenaltyPaymentsProcessing","fields":[{"name":"email","type":"string"},{"name":"payable_ref","type":"string"}]}`
		avroSchema := &avro.Schema{Definition: kafkaSchema}
		message := getConsumerMessage(avroSchema, penaltyPayment)
		mockFinancePayment := new(mockPenaltyFinancePayment)

		// When
		err := handleMessage(avroSchema, message, mockFinancePayment, cfg)

		// Then
		So(err, ShouldBeError, errors.New("error parsing the penalty-payments-processing avro encoded data: [End of file reached]"))
		mockFinancePayment.AssertNotCalled(t, "ProcessFinancialPenaltyPayment", penaltyPayment, e5PaymentID, cfg)
	})
}

func TestUnitHandleMessage_ProcessFinancialPenaltyPaymentFails(t *testing.T) {
	Convey("Handle message penalty payments processing fails", t, func() {
		// Given
		avroSchema := getAvroSchema()
		message := getConsumerMessage(avroSchema, penaltyPayment)
		mockFinancePayment := new(mockPenaltyFinancePayment)
		mockFinancePayment.On("ProcessFinancialPenaltyPayment", penaltyPayment, e5PaymentID, cfg).
			Return(errors.New("failed to create payment in E5"))

		// When
		err := handleMessage(avroSchema, message, mockFinancePayment, cfg)

		// Then
		So(err, ShouldBeError, errors.New("error processing financial penalty payment: [failed to create payment in E5]"))
		mockFinancePayment.AssertExpectations(t)
	})
}

func getAvroSchema() *avro.Schema {
	kafkaSchema := "{\"namespace\":\"uk.gov.companieshouse.financialpenalties\",\"type\":\"record\",\"doc\":\"thedetailsofthepenaltypaymentsbeingprocessed\",\"name\":\"PenaltyPaymentsProcessing\",\"fields\":[{\"name\":\"attempt\",\"type\":\"int\",\"default\":0,\"doc\":\"NumberofattemptstoretrypublishingthemessagetoKafkaTopic\"},{\"name\":\"company_code\",\"type\":\"string\"},{\"name\":\"customer_code\",\"type\":\"string\"},{\"name\":\"payment_id\",\"type\":\"string\"},{\"name\":\"external_payment_id\",\"type\":\"string\"},{\"name\":\"payment_reference\",\"type\":\"string\"},{\"name\":\"payment_amount\",\"type\":\"string\"},{\"name\":\"total_value\",\"type\":\"double\"},{\"name\":\"transaction_payments\",\"type\":{\"type\":\"array\",\"items\":{\"name\":\"transaction_payment\",\"type\":\"record\",\"fields\":[{\"name\":\"transaction_reference\",\"type\":\"string\"},{\"name\":\"value\",\"type\":\"double\"}]}}},{\"name\":\"card_type\",\"type\":\"string\"},{\"name\":\"email\",\"type\":\"string\"},{\"name\":\"payable_ref\",\"type\":\"string\"}]}"
	return &avro.Schema{Definition: kafkaSchema}
}

func getConsumerMessage(avroSchema *avro.Schema, penaltyPayment models.PenaltyPaymentsProcessing) *sarama.ConsumerMessage {
	avroBytes, _ := avroSchema.Marshal(penaltyPayment)

	return &sarama.ConsumerMessage{
		Value:     avroBytes,
		Topic:     "penalty-payments-processing",
		Partition: 0,
		Offset:    0,
	}
}
