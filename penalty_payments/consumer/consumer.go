package consumer

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	consumer "github.com/companieshouse/chs.go/kafka/consumer/cluster"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/handlers"
)

func Consume(cfg *config.Config) {
	kafkaConsumerConfig := &consumer.Config{
		BrokerAddr:   cfg.BrokerAddr,
		ZookeeperURL: cfg.ZookeeperURL,
		Topics:       []string{cfg.PenaltyPaymentsProcessingTopic},
	}
	log.Info("Starting kafka consumer", log.Data{
		"broker_addr":        kafkaConsumerConfig.BrokerAddr,
		"zookeeper_url":      kafkaConsumerConfig.ZookeeperURL,
		"topics":             kafkaConsumerConfig.Topics,
		"processing_timeout": kafkaConsumerConfig.ProcessingTimeout,
	})
	kafkaSchema, err := schema.Get(cfg.SchemaRegistryURL, cfg.PenaltyPaymentsProcessingTopic)
	if err != nil {
		kafkaSchemaError := fmt.Errorf("error getting penalty-payments-processing schema from schema registry: [%v]", err)
		panic(kafkaSchemaError)
	}
	avroSchema := &avro.Schema{
		Definition: kafkaSchema,
	}

	partitionConsumer := consumer.NewPartitionConsumer(kafkaConsumerConfig)

	if err := partitionConsumer.ConsumePartition(0, consumer.OffsetOldest); err != nil {
		log.Error(err)
	}

	defer func(partitionConsumer *consumer.PartitionConsumer) {
		err := partitionConsumer.Close()
		if err != nil {
			log.Error(err)
		}
	}(partitionConsumer)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	messages := partitionConsumer.Messages()
	penaltyFinancePayment := &handlers.PenaltyFinancePayment{}

	for {
		select {
		case <-c:
			log.Debug("Application terminating...")
			return
		case message := <-messages:
			if message != nil {
				err := handleMessage(avroSchema, message, penaltyFinancePayment)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}

}

func handleMessage(avroSchema *avro.Schema, message *sarama.ConsumerMessage, financePayment handlers.FinancePayment) error {
	var penaltyPayment models.PenaltyPaymentsProcessing
	var err = avroSchema.Unmarshal(message.Value, &penaltyPayment)
	if err != nil {
		return fmt.Errorf("error parsing the penalty-payments-processing avro encoded data: [%v]", err)
	}

	// this will be used for the PUON value in E5. it is referred to as paymentId in their spec. X is prefixed to it
	// so that it doesn't clash with other PUON's from different sources when finance produce their reports - namely
	// ones that begin with 'LP' which signify penalties that have been paid outside the digital service.
	e5PaymentID := "X" + penaltyPayment.PaymentID

	err = financePayment.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID)
	if err != nil {
		err = fmt.Errorf("error processing financial penalty payment: [%v]", err)
		log.Error(err, log.Data{"e5_payment_id": e5PaymentID, "customer_code": penaltyPayment.CustomerCode, "company_code": penaltyPayment.CompanyCode, "payable_ref": penaltyPayment.PayableRef})
		return err
	}

	return nil
}
