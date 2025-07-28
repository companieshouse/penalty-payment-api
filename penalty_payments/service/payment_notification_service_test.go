package service

import (
	"errors"
	"net/http"

	"testing"

	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var paymentInfo = validators.PaymentInformation{}

func TestUnitPaymentProcessingKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("Given the PaymentProcessingKafkaMessage is called", t, func() {
		Convey("When config is called with invalid config", func() {
			errMsg := "config is invalid"
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, errors.New(errMsg)
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := PaymentProcessingKafkaMessage(payableResource, &paymentInfo)

				So(err, ShouldResemble, errors.New("error getting config for penalty payments processing kafka message production: ["+errMsg+"]"))
			})
		})
		Convey("When config is called with valid config and invalid broker address", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetProducer := func(config *config.Config) (*producer.Producer, error) {
				return nil, errors.New("kafka: invalid configuration (You must provide at least one broker address)")
			}

			getConfig = mockedConfigGet
			getProducer = mockedGetProducer

			Convey("Then an error should be returned", func() {
				err := PaymentProcessingKafkaMessage(payableResource, &paymentInfo)

				So(err, ShouldResemble, errors.New("error creating penalty payments processing kafka producer: [kafka: invalid configuration (You must provide at least one broker address)]"))
			})
		})
		Convey("When config is called with valid config and valid broker config but invalid schema", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetProducer := func(config *config.Config) (*producer.Producer, error) {
				return &producer.Producer{}, nil
			}
			mockedGetSchema := func(url, schemaName string) (string, error) {
				return "", errors.New("get \"/subjects/penalty-payments-processing/versions/latest\": unsupported protocol scheme \"\"")
			}

			getConfig = mockedConfigGet
			getProducer = mockedGetProducer
			getSchema = mockedGetSchema

			Convey("Then an error should be returned", func() {
				err := PaymentProcessingKafkaMessage(payableResource, &paymentInfo)

				So(err, ShouldResemble, errors.New("error getting penalty payments processing schema from schema registry: [get \"/subjects/penalty-payments-processing/versions/latest\": unsupported protocol scheme \"\"]"))
			})
		})
		Convey("When config is called with valid config and valid broker config and valid schema", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetProducer := func(config *config.Config) (*producer.Producer, error) {
				return &producer.Producer{}, nil
			}
			mockedGetSchema := func(url, schemaName string) (string, error) {
				return "schema", nil
			}

			getConfig = mockedConfigGet
			getProducer = mockedGetProducer
			getSchema = mockedGetSchema

			Convey("Then an error should be returned", func() {
				err := PaymentProcessingKafkaMessage(payableResource, &paymentInfo)

				So(err.Error(), ShouldStartWith, "error preparing penalty payments processing kafka message with schema:")
			})
		})
	})
}

func TestUnitPreparePaymentProcessingKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	topic := "penalty-payments-processing"

	Convey("Given the PrepareKafkaMessage is called", t, func() {
		paymentsProcessingSchema, _ := schema.Get("chs.gov.uk", "penalty-payments-processing")
		producerSchema := avro.Schema{
			Definition: paymentsProcessingSchema,
		}

		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "LP", nil
		}

		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		Convey("When config is called with valid config and invalid company code", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
				return "", errors.New("error getting company code")
			}
			getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName

			Convey("Then an error should be returned", func() {
				_, err := preparePaymentProcessingKafkaMessage(producerSchema,
					payableResource, &paymentInfo, topic)

				So(err.Error(), ShouldEqual, "error getting company code")
			})
		})

		Convey("When config is called with invalid penalty ref", func() {
			mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
				return "", errors.New("error getting penalty ref type")
			}

			getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction

			Convey("Then an error should be returned", func() {
				_, err := preparePaymentProcessingKafkaMessage(
					producerSchema, payableResource, &paymentInfo, topic)

				// fix
				//So(err, ShouldResemble, errors.New("error getting penalty ref type"))
				So(err, ShouldNotBeNil)
			})
		})
		Convey("When config is called with no transaction items", func() {
			setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenRef)

			Convey("Then an error should be returned", func() {
				payableResourceNoItems := models.PayableResource{
					CustomerCode: customerCode,
					Transactions: []models.TransactionItem{},
				}

				_, err := preparePaymentProcessingKafkaMessage(
					producerSchema, payableResourceNoItems, &paymentInfo, topic)

				So(err.Error(), ShouldStartWith, "empty transactions list in payable resource:")
			})
		})
	})
}
