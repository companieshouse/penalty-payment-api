package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/handlers"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/consumer"
)

var consumerFunc = consumer.Consume

// SuperviseConsumer runs a consumer in a loop, restarting it if it exits unexpectedly
func SuperviseConsumer(ctx context.Context, name string, cfg *config.Config, penaltyFinancePayment *handlers.PenaltyFinancePayment, retry *resilience.ServiceRetry) {
	for {
		select {
		case <-ctx.Done():
			log.Info(fmt.Sprintf("Stopping supervise consumer: %s", name))
			return
		default:
			log.Info(fmt.Sprintf("Starting supervise consumer: %s", name))
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error(fmt.Errorf("panic recovered in supervise consumer %s: %v", name, r))
					}
				}()
				consumerFunc(cfg, penaltyFinancePayment, retry)
			}()

			log.Info(fmt.Sprintf("supervise consumer %s exited; restarting after delay", name))
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}
