package api

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/e5"
	. "github.com/smartystreets/goconvey/convey"
)

var companyNumber = "NI123456"
var companyCode = "LP"
var penaltyDetailsMap = &config.PenaltyDetailsMap{}
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
		},
	},
}
var transaction = e5.Transaction{
	CompanyCode:     "LP",
	TransactionType: "EU",
}
var page = e5.Page{
	Size:          1,
	TotalElements: 1,
	TotalPages:    1,
	Number:        1,
}
var e5TransactionsResponse = e5.GetTransactionsResponse{
	Page: page,
	Transactions: []e5.Transaction{
		1: transaction,
	},
}

func TestUnitAccountPenalties(t *testing.T) {
	Convey("error when no transactions provided", t, func() {
		_, responseType, err := AccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("penalties returned when valid transactions", t, func() {
		mockedGetTransactions := func(companyNumber string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("error when transactions cannot be found", t, func() {
		errGettingTransactions := errors.New("error getting transactions")
		mockedGetTransactions := func(companyNumber string, companyCode string, client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5.GetTransactionsResponse{}, errGettingTransactions
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldEqual, errGettingTransactions)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("error when generating transaction list fails", t, func() {
		errGeneratingTransactionList := errors.New("error generating transaction list from the e5 response: [error generating etag]")
		payableTransactionList := models.TransactionListResponse{}
		mockedGetTransactions := func(companyNumber string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}
		mockedGenerateTransactionList := func(e5Response *e5.GetTransactionsResponse, companyCode string,
			penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, error) {
			return &payableTransactionList, errors.New("error generating etag")
		}

		getTransactions = mockedGetTransactions
		generateTransactionList = mockedGenerateTransactionList

		listResponse, responseType, err := AccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldResemble, errGeneratingTransactionList)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("error when getConfig fails", t, func() {
		errGettingConfig := errors.New("error getting config")
		mockedGetConfig := func() (*config.Config, error) {
			return nil, errGettingConfig
		}

		getConfig = mockedGetConfig

		listResponse, responseType, err := AccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})
}
