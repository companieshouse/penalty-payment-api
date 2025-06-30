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
	for {
		select {
		case <-c:
			log.Debug("Application terminating...")
			return
		case message := <-messages:
			if message != nil {
				handleMessage(cfg, message)
			}
		}
	}

}

func handleMessage(cfg *config.Config, message *sarama.ConsumerMessage) {

	kafkaSchema, err := schema.Get(cfg.SchemaRegistryURL, cfg.PenaltyPaymentsProcessingTopic)
	if err != nil {
		err = fmt.Errorf("error getting schema from schema registry: [%v]", err)
	}
	avroSchema := &avro.Schema{
		Definition: kafkaSchema,
	}

	var penaltyPayment models.PenaltyPaymentsProcessing
	err = avroSchema.Unmarshal(message.Value, &penaltyPayment)
	if err != nil {
		return
	}

	log.Info("Processing penalty payment", log.Data{
		"attempt":               penaltyPayment.Attempt,
		"company_code":          penaltyPayment.CompanyCode,
		"customer_code":         penaltyPayment.CustomerCode,
		"payment_id":            penaltyPayment.PaymentId,
		"payment_external_id":   penaltyPayment.PaymentExternalId,
		"payment_reference":     penaltyPayment.PaymentReference,
		"payment_amount":        penaltyPayment.PaymentAmount,
		"total_value":           penaltyPayment.TotalValue,
		"transaction_reference": penaltyPayment.TransactionPayments[0].TransactionReference,
		"value":                 penaltyPayment.TransactionPayments[0].Value,
		"card_reference":        penaltyPayment.CardReference,
		"card_type":             penaltyPayment.CardType,
		"email":                 penaltyPayment.Email,
		"payable_ref":           penaltyPayment.PayableRef,
	})

}
