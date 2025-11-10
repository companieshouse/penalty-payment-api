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
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/testutils"
	"github.com/stretchr/testify/mock"
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
	// Start mock schema registry
	mockSchemaRegistry := startMockSchemaRegistry(t)
	defer mockSchemaRegistry.Close()

	// Start Kafka container
	kafkaContainer := testutils.NewKafkaContainer()
	kafkaContainer.Stop()
	kafkaContainer.Start()
	defer kafkaContainer.Stop()

	kafka3BrokerAddr := fmt.Sprintf("%s:%s", kafkaContainer.GetHost(), kafkaContainer.GetPort())

	// Config setup
	cfg := &config.Config{
		BrokerAddr:                             []string{kafka3BrokerAddr},
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

	// Setup mock with signal channel
	processed := make(chan struct{})
	mockFinancePayment := new(mockPenaltyFinancePayment)
	mockFinancePayment.On("ProcessFinancialPenaltyPayment", penaltyPayment, e5PaymentID, cfg, false).
		Return(nil).
		Run(func(args mock.Arguments) {
			log.Info("Mock ProcessFinancialPenaltyPayment called")
			close(processed)
		})

	// Start consumer
	done := make(chan struct{})
	go func() {
		Consume(cfg, mockFinancePayment, nil)
		close(done)
	}()

	// Give consumer time to start
	time.Sleep(5 * time.Second)

	// Produce message
	brokerAddrs := cfg.Kafka3BrokerAddr
	topic := cfg.PenaltyPaymentsProcessingTopic

	logContext := log.Data{
		"customer_code": penaltyPayment.CustomerCode,
		"payable_ref":   penaltyPayment.PayableRef,
		"broker_addrs":  brokerAddrs,
		"topic":         topic,
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokerAddrs, saramaConfig)
	require.NoError(t, err)
	defer producer.Close()

	avroBytes := getConsumerMessage(getTestAvroSchema(), penaltyPayment).Value
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(avroBytes),
	})
	log.Info("Sent test penalty-payments-processing message", log.Data{"partition": partition, "offset": offset}, logContext)
	require.NoError(t, err)

	// Wait for message to be processed or timeout
	select {
	case <-processed:
		log.Info("Message processed successfully", logContext)
	case <-time.After(2 * time.Minute):
		t.Fatal("Timeout waiting for message to be processed")
	}

	// Simulate shutdown
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(os.Interrupt)

	// Wait for graceful shutdown
	<-done

	// Assert expectations
	mockFinancePayment.AssertExpectations(t)
}
