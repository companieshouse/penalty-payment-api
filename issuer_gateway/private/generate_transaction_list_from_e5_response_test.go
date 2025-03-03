package private

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/e5"
	"github.com/companieshouse/penalty-payment-api/utils"

	. "github.com/smartystreets/goconvey/convey"
)

var companyCode = "LP"
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
			"S1": true,
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

var validSanctionsTransaction = e5.Transaction{
	CompanyCode:          utils.Sanctions,
	LedgerCode:           "E1",
	CustomerCode:         "12345678",
	TransactionReference: "P1234567",
	TransactionDate:      "2025-02-25",
	MadeUpDate:           "2025-02-12",
	Amount:               250,
	OutstandingAmount:    250,
	IsPaid:               false,
	TransactionType:      "1",
	TransactionSubType:   "S1",
	TypeDescription:      "CS01",
	DueDate:              "2025-03-26",
	AccountStatus:        ChsAccountStatus,
	DunningStatus:        Pen1DunningStatus,
}
var e5TransactionsResponseValidSanctions = e5.GetTransactionsResponse{
	Page: page,
	Transactions: []e5.Transaction{
		validSanctionsTransaction,
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

	Convey("transaction list successfully generated from E5 response - valid sanctions", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		transactionList, err := GenerateTransactionListFromE5Response(
			&e5TransactionsResponseValidSanctions, utils.Sanctions, &config.PenaltyDetailsMap{}, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "P1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "",
			IsPaid:          false,
			IsDCA:           false,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          "Failure to file a confirmation statement",
			PayableStatus:   "OPEN",
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("transaction list successfully generated from E5 response - valid sanctions with dunning status is dca", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		validSanctionsTransaction.DunningStatus = DcaDunningStatus
		e5TransactionsResponseValidSanctions = e5.GetTransactionsResponse{
			Page: page,
			Transactions: []e5.Transaction{
				validSanctionsTransaction,
			},
		}
		transactionList, err := GenerateTransactionListFromE5Response(
			&e5TransactionsResponseValidSanctions, utils.Sanctions, &config.PenaltyDetailsMap{}, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "P1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "",
			IsPaid:          false,
			IsDCA:           true,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          "Failure to file a confirmation statement",
			PayableStatus:   "CLOSED",
		}
		So(transactionListItem, ShouldResemble, expected)
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

func TestUnit_getPayableStatus(t *testing.T) {
	Convey("Get open payable status for late filing penalty", t, func() {
		type args struct {
			transaction *e5.Transaction
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty (valid)",
				args: args{transaction: &e5.Transaction{
					CompanyCode:          utils.LateFilingPenalty,
					LedgerCode:           "EW",
					CustomerCode:         "12345678",
					TransactionReference: "A1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               150,
					OutstandingAmount:    150,
					IsPaid:               false,
					TransactionType:      "1",
					TransactionSubType:   "EH",
					TypeDescription:      "Penalty Ltd Wel & Eng <=1m    LTDWA",
					DueDate:              "2025-03-26",
					AccountStatus:        ChsAccountStatus,
					DunningStatus:        Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount and not paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is dca, dunning status is not dca",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.transaction)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for late filing penalty", t, func() {
		type args struct {
			transaction *e5.Transaction
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty with outstanding amount is 0 and is paid (valid)",
				args: args{transaction: &e5.Transaction{
					CompanyCode:          utils.LateFilingPenalty,
					LedgerCode:           "EW",
					CustomerCode:         "12345678",
					TransactionReference: "A1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               150,
					OutstandingAmount:    0,
					IsPaid:               true,
					TransactionType:      "1",
					TransactionSubType:   "EH",
					TypeDescription:      "Penalty Ltd Wel & Eng <=1m    LTDWA",
					DueDate:              "2025-03-26",
					AccountStatus:        ChsAccountStatus,
					DunningStatus:        Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount is 0 and is paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 0,
					IsPaid:            true,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount less than 0 and is paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: -150,
					IsPaid:            true,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status and dunning status is dca",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     DcaDunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     DcaDunningStatus,
				}},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.transaction)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions", t, func() {
		type args struct {
			transaction *e5.Transaction
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions (valid)",
				args: args{transaction: &e5.Transaction{
					CompanyCode:          utils.Sanctions,
					LedgerCode:           "E1",
					CustomerCode:         "12345678",
					TransactionReference: "P1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               250,
					OutstandingAmount:    250,
					IsPaid:               false,
					TransactionType:      "1",
					TransactionSubType:   "S1",
					TypeDescription:      "CS01",
					DueDate:              "2025-03-26",
					AccountStatus:        ChsAccountStatus,
					DunningStatus:        Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HldAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     Pen2DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen2DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HldAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HldAccountStatus,
					DunningStatus:     Pen2DunningStatus,
				}},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.transaction)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions", t, func() {
		type args struct {
			transaction *e5.Transaction
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions with outstanding amount is 0 and is paid (valid)",
				args: args{transaction: &e5.Transaction{
					CompanyCode:          utils.Sanctions,
					LedgerCode:           "E1",
					CustomerCode:         "12345678",
					TransactionReference: "P1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               250,
					OutstandingAmount:    0,
					IsPaid:               true,
					TransactionType:      "1",
					TransactionSubType:   "S1",
					TypeDescription:      "CS01",
					DueDate:              "2025-03-26",
					AccountStatus:        ChsAccountStatus,
					DunningStatus:        Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount is 0 and is paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 0,
					IsPaid:            true,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount less than 0 and is paid",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: -250,
					IsPaid:            true,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status and dunning status is dca",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     DcaDunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     DcaDunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DcaAccountStatus,
					DunningStatus:     Pen3DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     ChsAccountStatus,
					DunningStatus:     Pen3DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HldAccountStatus,
					DunningStatus:     Pen3DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WdrAccountStatus,
					DunningStatus:     Pen1DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WdrAccountStatus,
					DunningStatus:     Pen2DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{transaction: &e5.Transaction{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WdrAccountStatus,
					DunningStatus:     Pen3DunningStatus,
				}},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.transaction)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}
