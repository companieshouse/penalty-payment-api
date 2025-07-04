package consumer

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/penalty-payment-api-core/models"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

type mockPenaltyFinancePayment struct {
	mock.Mock
}

func (m *mockPenaltyFinancePayment) ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string) error {
	args := m.Called(penaltyPayment, e5PaymentID)
	return args.Error(0)
}

func TestUnit_handleMessage(t *testing.T) {
	Convey("Process financial penalty payment", t, func() {
		// Given
		penaltyPayment := getPenaltyPayment()
		avroSchema := getAvroSchema()
		message := getConsumerMessage(avroSchema, penaltyPayment)
		mockFinancePayment := new(mockPenaltyFinancePayment)
		mockFinancePayment.On("ProcessFinancialPenaltyPayment", mock.Anything, mock.Anything).Return(nil)

		// When
		err := handleMessage(avroSchema, message, mockFinancePayment)

		// Then
		So(err, ShouldBeNil)
		mockFinancePayment.AssertCalled(t, "ProcessFinancialPenaltyPayment", penaltyPayment, "XKIYLUq1pRVuiLNA")
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

func getPenaltyPayment() models.PenaltyPaymentsProcessing {
	penaltyPayment := models.PenaltyPaymentsProcessing{
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
	return penaltyPayment
}
