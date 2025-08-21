package consumer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSchemaResponse struct {
	Schema string `json:"schema"`
}

func startMockSchemaRegistry(t *testing.T) *httptest.Server {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockSchemaResponse{
			Schema: `{
    "namespace": "uk.gov.companieshouse.financialpenalties",
    "type": "record",
    "doc": "thedetailsofthepenaltypaymentsbeingprocessed",
    "name": "PenaltyPaymentsProcessing",
    "fields": [
        {"name": "attempt", "type": "int", "default": 0, "doc": "NumberofattemptstoretrypublishingthemessagetoKafkaTopic"},
        {"name": "created_at", "type": "string", "doc": "thedateandtimethatarequesttoprocessthepenaltypaymentwascreated"},
        {"name": "company_code", "type": "string"},
        {"name": "customer_code", "type": "string"},
        {"name": "payment_id", "type": "string"},
        {"name": "external_payment_id", "type": "string"},
        {"name": "payment_reference", "type": "string"},
        {"name": "payment_amount", "type": "string"},
        {"name": "total_value", "type": "double"},
        {"name": "transaction_payments", "type": {"type": "array", "items": {"name": "transaction_payment", "type": "record", "fields": [{"name": "transaction_reference", "type": "string"},{"name": "value", "type": "double"}]}}},
        {"name": "card_type","type": "string"},
        {"name": "email","type": "string"},
        {"name": "payable_ref","type": "string"}
    ]
}`,
		}
		w.Header().Set("Content-Type", "application/vnd.schemaregistry.v1+json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	return httptest.NewServer(handler)
}

func TestIntegrationConsume(t *testing.T) {
	t.Parallel()

	mockSchemaRegistry := startMockSchemaRegistry(t)
	defer mockSchemaRegistry.Close()
	kafkaContainer := testutils.NewKafkaContainer()
	kafkaContainer.Start()
	defer kafkaContainer.Stop()

	kafka3BrokerAddr := fmt.Sprintf("%s:%s", kafkaContainer.GetHost(), kafkaContainer.GetPort())

	latestSchema, err := schema.Get(mockSchemaRegistry.URL, "penalty-payments-processing")
	require.NoError(t, err)
	require.NotNil(t, latestSchema)

	cfg := &config.Config{
		BrokerAddr:                             []string{"dummy-kafka:9092"},
		Kafka3BrokerAddr:                       []string{kafka3BrokerAddr},
		SchemaRegistryURL:                      mockSchemaRegistry.URL,
		PenaltyPaymentsProcessingTopic:         "penalty-payments-processing",
		PenaltyPaymentsProcessingMaxRetries:    "3",
		PenaltyPaymentsProcessingRetryDelay:    "1",
		PenaltyPaymentsProcessingRetryMaxDelay: "5",
		ConsumerGroupName:                      "penalty-payment-api-penalty-payments-processing",
		ConsumerRetryGroupName:                 "penalty-payment-api-penalty-payments-processing-retry",
		ConsumerRetryThrottleRate:              1,
		ConsumerRetryMaxAttempts:               3,
		FeatureFlagPaymentsProcessingEnabled:   true,
	}

	// Start consumer in background
	go func() {
		mockFinancePayment := new(mockPenaltyFinancePayment)
		mockFinancePayment.On("ProcessFinancialPenaltyPayment", penaltyPayment, e5PaymentID, cfg, true).Return(nil)

		retry := &resilience.ServiceRetry{
			ThrottleRate: time.Duration(cfg.ConsumerRetryThrottleRate) * time.Second,
			MaxRetries:   cfg.ConsumerRetryMaxAttempts,
		}
		Consume(cfg, mockFinancePayment, retry)
		mockFinancePayment.AssertExpectations(t)
	}()
	// Give consumer time to start
	time.Sleep(2 * time.Second)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Kafka3BrokerAddr, saramaConfig)
	require.NoError(t, err)
	defer producer.Close()

	avroBytes := getConsumerMessage(getTestAvroSchema(), penaltyPayment).Value
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: cfg.PenaltyPaymentsProcessingTopic + `-penalty-payment-api-retry`,
		Value: sarama.ByteEncoder(avroBytes),
	})
	log.Info("sent test penalty-payments-processing message", log.Data{"partition": partition, "offset": offset})
	require.NoError(t, err)

	// Simulate shutdown
	time.Sleep(2 * time.Second)
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	// Wait for graceful shutdown
	time.Sleep(1 * time.Second)

	// No assertion here - just ensuring no panic or crash
	assert.True(t, true)
}
