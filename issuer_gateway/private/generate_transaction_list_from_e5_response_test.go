package private

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/e5"

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

func TestUnit_getReason(t *testing.T) {
	Convey("Get reason", t, func() {
		type args struct {
			transaction *e5.Transaction
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing of accounts",
				args: args{transaction: &e5.Transaction{
					CompanyCode:        "LP",
					TransactionType:    "1",
					TransactionSubType: "Other",
				}},
				want: "Late filing of accounts",
			},
			{
				name: "Failure to file a confirmation statement",
				args: args{transaction: &e5.Transaction{
					CompanyCode:        "C1",
					TransactionType:    "1",
					TransactionSubType: "S1",
					TypeDescription:    "CS01",
				}},
				want: "Failure to file a confirmation statement",
			},
			{
				name: "Penalty",
				args: args{transaction: &e5.Transaction{
					CompanyCode:        "C1",
					TransactionType:    "1",
					TransactionSubType: "S1",
				}},
				want: "Penalty",
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getReason(tc.args.transaction)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}
