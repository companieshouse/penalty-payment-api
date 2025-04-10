package private

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

var now = time.Now().Truncate(time.Millisecond)

var sanctionsPenaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.Sanctions: {
			Description:        "Sanctions Penalty Payment",
			DescriptionId:      "penalty-sanctions",
			ClassOfPayment:     "penalty-sanctions",
			ResourceKind:       "penalty#sanctions",
			ProductType:        "penalty-sanctions",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		},
	},
}
var lfpPenaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.LateFilingPenalty: {
			Description:        "Late Filing Penalty",
			DescriptionId:      "late-filing-penalty",
			ClassOfPayment:     "penalty",
			ResourceKind:       "late-filing-penalty#late-filing-penalty",
			ProductType:        "late-filing-penalty",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		},
	},
}
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
			"S1": true,
		},
	},
}
var validSanctionsTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.Sanctions,
	LedgerCode:           "E1",
	CustomerCode:         "12345678",
	TransactionReference: "P1234567",
	TransactionDate:      "2025-02-25",
	MadeUpDate:           "2025-02-12",
	Amount:               250,
	OutstandingAmount:    250,
	IsPaid:               false,
	TransactionType:      SanctionsTransactionType,
	TransactionSubType:   SanctionsTransactionSubType,
	TypeDescription:      "CS01                                    ",
	DueDate:              "2025-03-26",
	AccountStatus:        CHSAccountStatus,
	DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
}
var validLFPTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.LateFilingPenalty,
	LedgerCode:           "EW",
	CustomerCode:         "12345678",
	TransactionReference: "A1234567",
	TransactionDate:      "2025-02-25",
	MadeUpDate:           "2025-02-12",
	Amount:               250,
	OutstandingAmount:    250,
	IsPaid:               false,
	TransactionType:      "1",
	TransactionSubType:   "EU",
	DueDate:              "2025-03-26",
	AccountStatus:        CHSAccountStatus,
	DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
}
var e5TransactionsResponseValidSanctions = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.Sanctions,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validSanctionsTransaction,
	},
}
var e5TransactionsResponseValidLFPTransaction = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.LateFilingPenalty,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validLFPTransaction,
	},
}

func TestUnitGenerateTransactionListFromE5Response(t *testing.T) {
	Convey("error when etag generator fails", t, func() {
		errorGeneratingEtag := errors.New("error generating etag")
		etagGenerator = func() (string, error) {
			return "", errorGeneratingEtag
		}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenalty, lfpPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldNotBeNil)
		So(transactionList, ShouldBeNil)
	})

	Convey("penalty list successfully generated from E5 response - penalty type EU", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenalty, lfpPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "A1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "late-filing-penalty#late-filing-penalty",
			IsPaid:          false,
			IsDCA:           false,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          LateFilingPenaltyReason,
			PayableStatus:   OpenPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("penalty list successfully generated from E5 response - penalty type Other", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "Other"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenalty, lfpPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "A1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "late-filing-penalty#late-filing-penalty",
			IsPaid:          false,
			IsDCA:           false,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "other",
			Reason:          LateFilingPenaltyReason,
			PayableStatus:   OpenPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("penalty list successfully generated from E5 response - valid lfp with dunning status is dca", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].DunningStatus = addTrailingSpacesToDunningStatus(DCADunningStatus)
		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenalty, lfpPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "A1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "late-filing-penalty#late-filing-penalty",
			IsPaid:          false,
			IsDCA:           true,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          LateFilingPenaltyReason,
			PayableStatus:   ClosedPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("penalty list successfully generated from E5 response - valid sanctions", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidSanctions, utils.Sanctions, sanctionsPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "P1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "penalty#sanctions",
			IsPaid:          false,
			IsDCA:           false,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          ConfirmationStatementReason,
			PayableStatus:   OpenPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("penalty list successfully generated from E5 response - valid sanctions with dunning status is dca", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidSanctions.AccountPenalties[0].DunningStatus = addTrailingSpacesToDunningStatus(DCADunningStatus)
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidSanctions, utils.Sanctions, sanctionsPenaltyDetailsMap, allowedTransactionMap)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "P1234567",
			Etag:            transactionListItem.Etag,
			Kind:            "penalty#sanctions",
			IsPaid:          false,
			IsDCA:           true,
			DueDate:         "2025-03-26",
			MadeUpDate:      "2025-02-12",
			TransactionDate: "2025-02-25",
			OriginalAmount:  250,
			Outstanding:     250,
			Type:            "penalty",
			Reason:          ConfirmationStatementReason,
			PayableStatus:   ClosedPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})
}

func TestUnit_getReason(t *testing.T) {
	Convey("Get reason", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing of accounts",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.LateFilingPenalty,
					TransactionType:    "1",
					TransactionSubType: "C1",
				}},
				want: LateFilingPenaltyReason,
			},
			{
				name: "Failure to file a confirmation statement",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.Sanctions,
					TransactionType:    SanctionsTransactionType,
					TransactionSubType: SanctionsTransactionSubType,
					TypeDescription:    "CS01                                    ",
				}},
				want: ConfirmationStatementReason,
			},
			{
				name: "Penalty",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.Sanctions,
					TransactionType:    SanctionsTransactionType,
					TransactionSubType: SanctionsTransactionSubType,
					TypeDescription:    "P&S Penalty                             ",
				}},
				want: PenaltyReason,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getReason(tc.args.penalty)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func addTrailingSpacesToDunningStatus(dunningStatus string) string {
	return fmt.Sprintf("%s%s", dunningStatus, "        ")
}

func TestUnit_getPayableStatus(t *testing.T) {
	Convey("Get open payable status for late filing penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty (valid)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
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
					AccountStatus:        CHSAccountStatus,
					DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount and not paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.penalty)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for late filing penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
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
					AccountStatus:        CHSAccountStatus,
					DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount is 0 and is paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 0,
					IsPaid:            true,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount less than 0 and is paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: -150,
					IsPaid:            true,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA (no trailing spaces)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     DCADunningStatus,
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is DCA",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is DCA",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is DCA",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is IPEN1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     "IPEN1",
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is IPEN2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     "IPEN2",
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is CAN",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.LateFilingPenalty,
					OutstandingAmount: 150,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     "CAN",
				}},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.penalty)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions (valid)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:          utils.Sanctions,
					LedgerCode:           "E1",
					CustomerCode:         "12345678",
					TransactionReference: "P1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               250,
					OutstandingAmount:    250,
					IsPaid:               false,
					TransactionType:      SanctionsTransactionType,
					TransactionSubType:   SanctionsTransactionSubType,
					TypeDescription:      "CS01                                    ",
					DueDate:              "2025-03-26",
					AccountStatus:        CHSAccountStatus,
					DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.penalty)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:          utils.Sanctions,
					LedgerCode:           "E1",
					CustomerCode:         "12345678",
					TransactionReference: "P1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               250,
					OutstandingAmount:    0,
					IsPaid:               true,
					TransactionType:      SanctionsTransactionType,
					TransactionSubType:   SanctionsTransactionSubType,
					TypeDescription:      "CS01                                    ",
					DueDate:              "2025-03-26",
					AccountStatus:        CHSAccountStatus,
					DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount is 0 and is paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 0,
					IsPaid:            true,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount less than 0 and is paid",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: -250,
					IsPaid:            true,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(DCADunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     DCAAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     CHSAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     HLDAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN1DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN2DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:       utils.Sanctions,
					OutstandingAmount: 250,
					IsPaid:            false,
					AccountStatus:     WDRAccountStatus,
					DunningStatus:     addTrailingSpacesToDunningStatus(PEN3DunningStatus),
				}},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := getPayableStatus(tc.args.penalty)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}
