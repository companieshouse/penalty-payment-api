package private

import (
	"errors"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/e5"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var companyCode = "LP"
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
		},
	},
}
var euTransaction = e5.Transaction{
	CompanyCode:        "LP",
	TransactionType:    "1",
	TransactionSubType: "EU",
}
var page = e5.Page{
	Size:          1,
	TotalElements: 1,
	TotalPages:    1,
	Number:        1,
}
var e5TransactionsResponseEu = e5.GetTransactionsResponse{
	Page: page,
	Transactions: []e5.Transaction{
		euTransaction,
	},
}

func TestUnitGenerateTransactionListFromE5Response(t *testing.T) {
	Convey("error when etag generator fails", t, func() {
		errorGeneratingEtag := errors.New("error generating etag")
		etagGenerator = func() (string, error) {
			return "", errorGeneratingEtag
		}

		transactionList, err := GenerateTransactionListFromE5Response(
			&e5TransactionsResponseEu, companyCode, &config.PenaltyDetailsMap{}, allowedTransactionMap)
		So(err, ShouldNotBeNil)
		So(transactionList, ShouldBeNil)
	})

	Convey("transaction list successfully generated from E5 response - transaction type EU", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		transactionList, err := GenerateTransactionListFromE5Response(
			&e5TransactionsResponseEu, companyCode, &config.PenaltyDetailsMap{}, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
	})

	Convey("transaction list successfully generated from E5 response - transaction type Other", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		otherTransaction := e5.Transaction{
			CompanyCode:        "LP",
			TransactionType:    "1",
			TransactionSubType: "Other",
		}
		e5TransactionsResponseOther := e5.GetTransactionsResponse{
			Page: page,
			Transactions: []e5.Transaction{
				otherTransaction,
			},
		}

		transactionList, err := GenerateTransactionListFromE5Response(
			&e5TransactionsResponseOther, companyCode, &config.PenaltyDetailsMap{}, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
	})
}
