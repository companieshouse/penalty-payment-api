package service

import (
	"errors"
	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var req = &http.Request{}
var penaltyDetailsMap = &config.PenaltyDetailsMap{}
var allowedTransactionsMap = &models.AllowedTransactionMap{}
var transactionItem = models.TransactionItem{
	TransactionID: "1235689",
}
var payableResource = models.PayableResource{
	CompanyNumber: companyNumber,
	Transactions:  []models.TransactionItem{transactionItem},
}

func TestUnitSendEmailKafkaMessage(t *testing.T) {
	Convey("Given the SendEmailKafkaMessage is called", t, func() {
		Convey("When config is called with invalid config", func() {
			errMsg := "config is invalid"
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, errors.New(errMsg)
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error getting config for kafka message production: ["+errMsg+"]"))
			})
		})
		Convey("When config is called with valid config and invalid broker address", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error creating kafka producer: [kafka: invalid configuration (You must provide at least one broker address)]"))
			})
		})
		Convey("When config is called with valid config and valid broker config but invalid schema", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetProducer := func(config *config.Config) (*producer.Producer, error) {
				return &producer.Producer{}, nil
			}

			getConfig = mockedConfigGet
			getProducer = mockedGetProducer

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error getting schema from schema registry: [Get \"/subjects/email-send/versions/latest\": unsupported protocol scheme \"\"]"))
			})
		})
		Convey("When config is called with valid config and valid broker config and valid schema", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetProducer := func(config *config.Config) (*producer.Producer, error) {
				return &producer.Producer{}, nil
			}
			mockedGetSchema := func(url string) (string, error) {
				return "schema", nil
			}

			getConfig = mockedConfigGet
			getProducer = mockedGetProducer
			getSchema = mockedGetSchema

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err.Error(), ShouldStartWith, "error preparing kafka message with schema: [error getting company name: [")
			})
		})
	})
}

func TestUnitPrepareKafkaMessage(t *testing.T) {
	Convey("Given the PrepareKafkaMessage is called", t, func() {
		emailSendSchema, _ := schema.Get("chs.gov.uk", ProducerSchemaName)
		producerSchema := avro.Schema{
			Definition: emailSendSchema,
		}

		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "LP", nil
		}

		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		Convey("When config is called with invalid config", func() {
			errMsg := "config is invalid"
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, errors.New(errMsg)
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error getting config: ["+errMsg+"]"))
			})
		})
		Convey("When config is called with valid config and invalid company number", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err.Error(), ShouldStartWith, "error getting company name: [")
			})
		})
		Convey("When config is called with valid config and valid company number but invalid transaction", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err.Error(), ShouldStartWith, "error getting transaction for penalty: [")
			})
		})
		Convey("When config is called with valid config and valid company number and valid transaction but invalid madeUpDate", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetTransactionForPenalty := func(companyNumber, penaltyNumber string, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListItem, error) {

				return &models.TransactionListItem{ID: "1235689"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getTransactionForPenalty = mockedGetTransactionForPenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error parsing made up date: [parsing time \"\" as \"2006-01-02\": cannot parse \"\" as \"2006\"]"))
			})
		})
		Convey("When config is called with valid config and valid company number and valid transaction but invalid penalty date", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetTransactionForPenalty := func(companyNumber, penaltyNumber string, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListItem, error) {

				return &models.TransactionListItem{ID: "1235689", MadeUpDate: "2006-01-02"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getTransactionForPenalty = mockedGetTransactionForPenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error parsing penalty date: [parsing time \"\" as \"2006-01-02\": cannot parse \"\" as \"2006\"]"))
			})
		})
		Convey("When config is called with valid config and valid company number and valid transaction and valid penalty date but unknown type", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetTransactionForPenalty := func(companyNumber, penaltyNumber string, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListItem, error) {

				return &models.TransactionListItem{ID: "1235689", MadeUpDate: "2006-01-02", TransactionDate: "2006-01-02"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getTransactionForPenalty = mockedGetTransactionForPenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap)

				So(err, ShouldResemble, errors.New("error marshalling email send message: [Unknown type name: ]"))
			})
		})
	})
}
