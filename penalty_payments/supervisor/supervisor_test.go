package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/handlers"
)

var mockConsumerFunc = func(cfg *config.Config, penaltyFinancePayment handlers.FinancePayment, retry *resilience.ServiceRetry) {
	panic("simulated panic")
}

func TestUnitSuperviseConsumer_PanicRecoveryAndShutdown(t *testing.T) {
	// Override the consumer.Consume function with the mock
	original := consumerFunc
	consumerFunc = mockConsumerFunc
	defer func() {
		consumerFunc = original
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{}
	penaltyFinancePayment := &handlers.PenaltyFinancePayment{}
	retry := &resilience.ServiceRetry{}

	done := make(chan struct{})

	go func() {
		SuperviseConsumer(ctx, "test-consumer", cfg, penaltyFinancePayment, retry)
		close(done)
	}()

	// Let the consumer run briefly
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("SuperviseConsumer did not exit after context cancellation")
	}
}
