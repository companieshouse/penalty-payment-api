package service

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/e5"
	"github.com/companieshouse/penalty-payment-api/utils"

	. "github.com/smartystreets/goconvey/convey"
)

var companyNumber = "NI123456"
var companyCode = utils.LateFilingPenalty
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
		},
	},
}
var transaction = e5.Transaction{
	CompanyCode:     utils.LateFilingPenalty,
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

func TestUnitGetPenalties(t *testing.T) {
	Convey("error when no transactions provided", t, func() {
		_, responseType, err := GetPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("penalties returned when valid transactions", t, func() {
		mockedGetTransactions := func(companyNumber string, companyCode string,
			penaltyDetailsMap *config.PenaltyDetailsMap, client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := GetPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("error when transactions cannot be found", t, func() {
		errGettingTransactions := errors.New("error getting transactions")
		mockedGetTransactions := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5.GetTransactionsResponse{}, errGettingTransactions
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := GetPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldEqual, errGettingTransactions)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})
}
