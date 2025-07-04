package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/companieshouse/chs.go/avro"
	"github.com/companieshouse/chs.go/avro/schema"
	"github.com/companieshouse/chs.go/kafka/producer"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var req = &http.Request{}
var penaltyDetailsMap = &config.PenaltyDetailsMap{}
var allowedTransactionsMap = &models.AllowedTransactionMap{}
var customerCode = "NI123456"
var transactionItem = models.TransactionItem{
	PenaltyRef: "A1234567",
}
var chsUrl = "chs_url"
var payableResource = models.PayableResource{
	CustomerCode: customerCode,
	Transactions: []models.TransactionItem{transactionItem},
}

func TestUnitSendEmailKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("Given the SendEmailKafkaMessage is called", t, func() {
		Convey("When config is called with invalid config", func() {
			errMsg := "config is invalid"
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, errors.New(errMsg)
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

				So(err, ShouldResemble, errors.New("error getting config for kafka message production: ["+errMsg+"]"))
			})
		})
		Convey("When config is called with valid config and invalid broker address", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

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
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

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
				err := SendEmailKafkaMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

				So(err.Error(), ShouldStartWith, "error preparing kafka message with schema: [error getting company name: [")
			})
		})
	})
}

func setGetPenaltyRefTypeFromTransactionMock(penaltyRefType string) {
	mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return penaltyRefType, nil
	}
	getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction
}

func TestUnitPrepareKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("Given the PrepareKafkaMessage is called", t, func() {
		emailSendSchema, _ := schema.Get("chs.gov.uk", ProducerSchemaName)
		producerSchema := avro.Schema{
			Definition: emailSendSchema,
		}

		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "LP", nil
		}

		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		Convey("When getting company name fails config", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(
					producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil, chsUrl)

				So(err.Error(), ShouldStartWith, "error getting company name: [")
			})
		})
		Convey("When getting company code fails", func() {
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
				_, err := prepareKafkaMessage(
					producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil, chsUrl)

				So(err.Error(), ShouldEqual, "error getting company code")
			})
		})
		Convey("When getting penalty ref fails", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
				return "", errors.New("error getting penalty ref type")
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(
					producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil, chsUrl)

				So(err, ShouldResemble, errors.New("error getting penalty ref type"))
			})
		})
		Convey("When payable resource has no transaction items", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenRef)

			mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)

			Convey("Then an error should be returned", func() {
				payableResourceNoItems := models.PayableResource{
					CustomerCode: customerCode,
					Transactions: []models.TransactionItem{},
				}

				_, err := prepareKafkaMessage(
					producerSchema, payableResourceNoItems, req, penaltyDetailsMap, allowedTransactionsMap, mockApDaoSvc, chsUrl)

				So(err.Error(), ShouldStartWith, "empty transactions list in payable resource:")
			})
		})
		Convey("When getting transaction for provided penalty ref fails", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenRef)

			mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)

			Convey("Then an error should be returned", func() {
				mockApDaoSvc.EXPECT().GetAccountPenalties(gomock.Any(), gomock.Any()).Return(nil, nil)

				_, err := prepareKafkaMessage(
					producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, mockApDaoSvc, chsUrl)

				So(err.Error(), ShouldStartWith, "error getting transaction for penalty: [")
			})
		})
		Convey("When transaction's madeUpDate is invalid", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetPayablePenalty := func(penaltyRefType, customerCode, companyCode string, t models.TransactionItem, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionItem, error) {

				return &models.TransactionItem{PenaltyRef: "A1234567", Reason: "Late filing of accounts"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getPayablePenalty = mockedGetPayablePenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(
					producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil, chsUrl)

				So(err, ShouldResemble, errors.New("error parsing made up date: [parsing time \"\" as \"2006-01-02\": cannot parse \"\" as \"2006\"]"))
			})
		})
		Convey("When penalty type is unknown type", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetPayablePenalty := func(penaltyRefType, customerCode, companyCode string, t models.TransactionItem, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionItem, error) {

				return &models.TransactionItem{
					PenaltyRef: "A123567",
					MadeUpDate: "2006-01-02",
					Reason:     "Late filing of accounts"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getPayablePenalty = mockedGetPayablePenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareKafkaMessage(producerSchema, payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil, chsUrl)

				So(err, ShouldResemble, errors.New("error marshalling email send message: [Unknown type name: ]"))
			})
		})
	})
}
