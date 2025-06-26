package service

import (
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
)

var getConfig = config.Get
var getProducer = func(config *config.Config) (*producer.Producer, error) {
	return producer.New(&producer.Config{Acks: &producer.WaitForAll, BrokerAddrs: config.BrokerAddr})
}
var getSchema = func(url, schemaName string) (string, error) {
	return schema.Get(url, schemaName)
}

var getCompanyCodeFromTransaction = utils.GetCompanyCodeFromTransaction
var getPenaltyRefTypeFromTransaction = utils.GetPenaltyRefTypeFromTransaction
var getCompanyName = GetCompanyName
var getPayablePenalty = api.PayablePenalty
