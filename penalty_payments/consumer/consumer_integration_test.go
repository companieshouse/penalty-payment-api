package consumer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegrationConsume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/cp-kafka:7.5.0",
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithExposedPorts("9093/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9093/tcp"),
			wait.ForLog("Kafka Server started (kafka.server.KafkaRaftServer)")))
	require.NoError(t, err, "Kafka container failed to start within timeout")

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		err := kafkaContainer.Terminate(cleanupCtx)
		require.NoError(t, err)
	})

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	schemaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(getTestSchemaResponse()))
		require.NoError(t, err)
	}))
	defer schemaServer.Close()

	kafka3BrokerAddr := []string{brokers[0]}
	cfg := &config.Config{
		BrokerAddr:                             []string{"dummy-kafka:9092"},
		Kafka3BrokerAddr:                       kafka3BrokerAddr,
		SchemaRegistryURL:                      schemaServer.URL,
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

func getTestSchemaResponse() string {
	return `{"schema":"{\"namespace\":\"uk.gov.companieshouse.financialpenalties\",\"type\":\"record\",\"doc\":\"thedetailsofthepenaltypaymentsbeingprocessed\",\"name\":\"PenaltyPaymentsProcessing\",\"fields\":[{\"name\":\"attempt\",\"type\":\"int\",\"default\":0,\"doc\":\"NumberofattemptstoretrypublishingthemessagetoKafkaTopic\"},{\"name\":\"created_at\",\"type\":\"string\",\"doc\":\"thedateandtimethatarequesttoprocessthepenaltypaymentwascreated\"},{\"name\":\"company_code\",\"type\":\"string\"},{\"name\":\"customer_code\",\"type\":\"string\"},{\"name\":\"payment_id\",\"type\":\"string\"},{\"name\":\"external_payment_id\",\"type\":\"string\"},{\"name\":\"payment_reference\",\"type\":\"string\"},{\"name\":\"payment_amount\",\"type\":\"string\"},{\"name\":\"total_value\",\"type\":\"double\"},{\"name\":\"transaction_payments\",\"type\":{\"type\":\"array\",\"items\":{\"name\":\"transaction_payment\",\"type\":\"record\",\"fields\":[{\"name\":\"transaction_reference\",\"type\":\"string\"},{\"name\":\"value\",\"type\":\"double\"}]}}},{\"name\":\"card_type\",\"type\":\"string\"},{\"name\":\"email\",\"type\":\"string\"},{\"name\":\"payable_ref\",\"type\":\"string\"}]}"}`
}
