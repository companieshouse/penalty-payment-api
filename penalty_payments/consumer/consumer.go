package consumer

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	consumer "github.com/companieshouse/chs.go/kafka/consumer/cluster"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
)

func Consume(cfg *config.Config, penaltyFinancePayment api.FinancePayment, retry *resilience.ServiceRetry) {
	avroSchema := getAvroSchema(cfg)
	topic := cfg.PenaltyPaymentsProcessingTopic
	resilienceHandler := resilience.NewHandler(topic, cfg.Namespace(), retry, getProducer(cfg), avroSchema)

	consumerGroupName := cfg.ConsumerGroupName
	isRetry := retry != nil
	if isRetry {
		topic = resilienceHandler.GetRetryTopicName()
		consumerGroupName = cfg.ConsumerRetryGroupName
	}
	consumerConfig := &consumer.Config{
		BrokerAddr: cfg.Kafka3BrokerAddr,
		Topics:     []string{topic},
	}
	log.Info("Starting Kafka3 consumer with resilience", log.Data{
		"consumer_config": consumerConfig,
		"retry":           retry,
	})

	var resetOffset bool
	consumerGroupConfig := &consumer.GroupConfig{
		GroupName:   consumerGroupName,
		ResetOffset: resetOffset,
	}

	groupConsumer := consumer.NewConsumerGroup(consumerConfig)

	if err := groupConsumer.JoinGroup(consumerGroupConfig); err != nil {
		log.Error(err)
	}

	defer func(groupConsumer *consumer.GroupConsumer) {
		err := groupConsumer.Close()
		if err != nil {
			log.Error(err)
		}
	}(groupConsumer)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	messages := groupConsumer.Messages()

	for {
		select {
		case <-c:
			log.Debug("Application terminating...")
			return
		case message := <-messages:
			if message != nil {
				err := handleMessage(avroSchema, message, penaltyFinancePayment, cfg, resilienceHandler, isRetry)
				if err != nil {
					log.Error(err)
				} else {
					groupConsumer.MarkOffset(message, "")
				}
			}
		}
	}

}

func handleMessage(avroSchema *avro.Schema, message *sarama.ConsumerMessage, financePayment api.FinancePayment,
	cfg *config.Config, resilience *resilience.Resilience, isRetry bool) error {
	log.Debug("Received message", log.Data{
		"message":  message,
		"is_retry": isRetry,
	})
	var penaltyPayment models.PenaltyPaymentsProcessing
	var err = avroSchema.Unmarshal(message.Value, &penaltyPayment)
	if err != nil {
		return fmt.Errorf("error parsing the penalty-payments-processing avro encoded data: [%v]", err)
	}

	// this will be used for the PUON value in E5. it is referred to as paymentId in their spec. X is prefixed to it
	// so that it doesn't clash with other PUON's from different sources when finance produce their reports - namely
	// ones that begin with 'LP' which signify penalties that have been paid outside the digital service.
	e5PaymentID := "X" + penaltyPayment.PaymentID

	logContext := log.Data{
		"customer_code": penaltyPayment.CustomerCode,
		"company_code":  penaltyPayment.CompanyCode,
		"payable_ref":   penaltyPayment.PayableRef,
		"e5_payment_id": e5PaymentID,
		"is_retry":      isRetry,
	}
	log.Debug("Consumer handle message - BEFORE Financial penalty payment processing", log.Data{
		"Topic":     message.Topic,
		"Partition": message.Partition,
		"Offset":    message.Offset,
	}, logContext)
	err = financePayment.ProcessFinancialPenaltyPayment(penaltyPayment, e5PaymentID, cfg, isRetry)
	if err != nil {
		err = fmt.Errorf("error processing financial penalty payment: [%v]", err)
		log.Error(err, logContext)
		return resilience.HandleError(err, message.Offset, &penaltyPayment)
	}

	return nil
}

func getAvroSchema(cfg *config.Config) *avro.Schema {
	kafkaSchema, err := schema.Get(cfg.SchemaRegistryURL, cfg.PenaltyPaymentsProcessingTopic)
	if err != nil {
		kafkaSchemaError := fmt.Errorf("error getting penalty-payments-processing schema from schema registry: [%v]", err)
		panic(kafkaSchemaError)
	}
	avroSchema := &avro.Schema{
		Definition: kafkaSchema,
	}
	return avroSchema
}

func getProducer(cfg *config.Config) *producer.Producer {
	kafkaProducerConfig := &producer.Config{Acks: &producer.WaitForAll, BrokerAddrs: cfg.Kafka3BrokerAddr}
	syncProducer, err := producer.New(kafkaProducerConfig)
	if err != nil {
		log.Error(fmt.Errorf("error initialising Kafka3 producer for resilience: %s", err), log.Data{
			"producer_config": kafkaProducerConfig,
		})
	}
	return syncProducer
}
