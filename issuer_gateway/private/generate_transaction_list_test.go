package private

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	. "github.com/smartystreets/goconvey/convey"
)

var now = time.Now().Truncate(time.Millisecond)
var yesterday = time.Now().AddDate(0, 0, -1).Truncate(time.Millisecond)

var sanctionsPenaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.SanctionsCompanyCode: {
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
var sanctionsRoePenaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.SanctionsCompanyCode: {
			Description:        "Overseas Entity Penalty Payment",
			DescriptionId:      "penalty-sanctions",
			ClassOfPayment:     "penalty-sanctions",
			ResourceKind:       "penalty#sanctions",
			ProductType:        "penalty-sanctions",
			EmailReceivedAppId: "penalty-payment-api.sanctions_roe_penalty_payment_received_email",
			EmailMsgType:       "sanctions_roe_penalty_payment_received_email",
		},
	},
}
var lfpPenaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.LateFilingPenaltyCompanyCode: {
			Description:        "Late Filing Penalty",
			DescriptionId:      "late-filing-penalty",
			ClassOfPayment:     "penalty-lfp",
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
			"A2": true,
		},
	},
}
var validSanctionsTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.SanctionsCompanyCode,
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
var validSanctionsRoeFailureToUpdateTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.SanctionsCompanyCode,
	LedgerCode:           "FU",
	CustomerCode:         "OE123456",
	TransactionReference: "U1234567",
	TransactionDate:      "2025-02-25",
	MadeUpDate:           "2025-02-12",
	Amount:               250,
	OutstandingAmount:    250,
	IsPaid:               false,
	TransactionType:      SanctionsTransactionType,
	TransactionSubType:   SanctionsRoeFailureToUpdateTransactionSubType,
	TypeDescription:      "PENU                                    ",
	DueDate:              "2025-03-26",
	AccountStatus:        CHSAccountStatus,
	DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
}
var validLFPTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.LateFilingPenaltyCompanyCode,
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
var validLFPUnpaidCostsTransaction = models.AccountPenaltiesDataDao{
	CompanyCode:          utils.LateFilingPenaltyCompanyCode,
	LedgerCode:           "EW",
	CustomerCode:         "12345678",
	TransactionReference: "F1",
	TransactionDate:      "2025-02-25",
	MadeUpDate:           "2025-02-12",
	Amount:               250,
	OutstandingAmount:    250,
	IsPaid:               false,
	TransactionType:      "2",
	TransactionSubType:   "EU",
	DueDate:              "2025-03-26",
	AccountStatus:        CHSAccountStatus,
	DunningStatus:        addTrailingSpacesToDunningStatus(PEN1DunningStatus),
}
var e5TransactionsResponseValidSanctions = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.SanctionsCompanyCode,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validSanctionsTransaction,
	},
}
var e5TransactionsResponseValidRoe = models.AccountPenaltiesDao{
	CustomerCode: "OE123456",
	CompanyCode:  utils.SanctionsCompanyCode,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validSanctionsRoeFailureToUpdateTransaction,
	},
}
var e5TransactionsResponseValidLFPTransaction = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.LateFilingPenaltyCompanyCode,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validLFPTransaction,
	},
}
var e5TransactionsResponseValidLFPWithUnpaidCostsTransaction = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.LateFilingPenaltyCompanyCode,
	CreatedAt:    &now,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		validLFPTransaction,
		validLFPUnpaidCostsTransaction,
	},
}

var cfg = config.Config{}

func TestUnitGenerateTransactionListFromE5Response(t *testing.T) {
	Convey("error when first etag generator fails", t, func() {
		mockedEtagGenerator := func() (string, error) {
			return "", errors.New("error generating etag")
		}
		etagGenerator = mockedEtagGenerator

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
		So(err.Error(), ShouldStartWith, "error generating etag")
		So(transactionList, ShouldBeNil)
	})

	Convey("error when first etag generator succeeds but second etag generator fails", t, func() {
		callCount := 0
		mockedEtagGenerator := func() (string, error) {
			callCount++
			if callCount == 2 {
				return "", errors.New("error generating etag")
			}
			return "ABCDE", nil
		}

		etagGenerator = mockedEtagGenerator

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
		So(err.Error(), ShouldStartWith, "error generating etag")
		So(transactionList, ShouldBeNil)
	})

	Convey("penalty list successfully generated from E5 response - unpaid costs", t, func() {
		mockedEtagGenerator := func() (string, error) {
			return "ABCDE", nil
		}
		etagGenerator = mockedEtagGenerator

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPWithUnpaidCostsTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 2)
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
			PayableStatus:   ClosedPayableStatus,
		}
		So(transactionListItem, ShouldResemble, expected)
	})

	Convey("penalty list successfully generated from E5 response - penalty type EU", t, func() {
		mockedEtagGenerator := func() (string, error) {
			return "ABCDE", nil
		}
		etagGenerator = mockedEtagGenerator

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
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

	Convey("penalty list successfully generated from E5 response - penalty type EU", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidLFPTransaction.AccountPenalties[0].TransactionSubType = "EU"
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
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
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
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
			PayableStatus:   ClosedPayableStatus,
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
			&e5TransactionsResponseValidLFPTransaction, utils.LateFilingPenaltyCompanyCode, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg)
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
			&e5TransactionsResponseValidSanctions, utils.SanctionsCompanyCode, sanctionsPenaltyDetailsMap, allowedTransactionMap, &cfg)
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

	Convey("penalty list successfully generated from E5 response - valid sanctions ROE", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidRoe, utils.SanctionsCompanyCode, sanctionsRoePenaltyDetailsMap, allowedTransactionMap, &cfg)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "U1234567",
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
			Reason:          SanctionsRoeFailureToUpdateReason,
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
			&e5TransactionsResponseValidSanctions, utils.SanctionsCompanyCode, sanctionsPenaltyDetailsMap, allowedTransactionMap, &cfg)
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

	Convey("penalty list successfully generated from E5 response - valid sanctions ROE with dunning status is dca", t, func() {
		etag := "ABCDE"
		etagGenerator = func() (string, error) {
			return etag, nil
		}

		e5TransactionsResponseValidRoe.AccountPenalties[0].DunningStatus = addTrailingSpacesToDunningStatus(DCADunningStatus)
		transactionList, err := GenerateTransactionListFromAccountPenalties(
			&e5TransactionsResponseValidRoe, utils.SanctionsCompanyCode, sanctionsRoePenaltyDetailsMap, allowedTransactionMap, &cfg)
		So(err, ShouldBeNil)
		So(transactionList, ShouldNotBeNil)
		transactionListItems := transactionList.Items
		So(len(transactionListItems), ShouldEqual, 1)
		transactionListItem := transactionListItems[0]
		expected := models.TransactionListItem{
			ID:              "U1234567",
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
			Reason:          SanctionsRoeFailureToUpdateReason,
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
					CompanyCode:        utils.LateFilingPenaltyCompanyCode,
					TransactionType:    "1",
					TransactionSubType: "C1",
				}},
				want: LateFilingPenaltyReason,
			},
			{
				name: "Failure to file a confirmation statement",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    SanctionsTransactionType,
					TransactionSubType: SanctionsTransactionSubType,
					TypeDescription:    "CS01                                    ",
				}},
				want: ConfirmationStatementReason,
			},
			{
				name: "Failure to update the Register of Overseas Entities",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    SanctionsTransactionType,
					TransactionSubType: SanctionsRoeFailureToUpdateTransactionSubType,
				}},
				want: SanctionsRoeFailureToUpdateReason,
			},
			{
				name: "Penalty",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
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
	return fmt.Sprintf("%-12s", dunningStatus)
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
				args: args{penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount and not paid",
				args: args{
					penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN1",
				args: args{
					penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN2",
				args: args{
					penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN3",
				args: args{
					penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN1",
				args: args{
					penalty: createLateFilingPenalty(false, 150, DCADunningStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN2",
				args: args{
					penalty: createLateFilingPenalty(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN3",
				args: args{
					penalty: createLateFilingPenalty(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN1",
				args: args{
					penalty: createLateFilingPenalty(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN2",
				args: args{
					penalty: createLateFilingPenalty(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN3",
				args: args{
					penalty: createLateFilingPenalty(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN1",
				args: args{
					penalty: createLateFilingPenalty(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN2",
				args: args{
					penalty: createLateFilingPenalty(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN3",
				args: args{
					penalty: createLateFilingPenalty(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

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
				args: args{penalty: createLateFilingPenalty(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount less than 0 and is paid",
				args: args{penalty: createLateFilingPenalty(true, -150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA",
				args: args{penalty: createLateFilingPenalty(false, 150, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA (no trailing spaces)",
				args: args{penalty: createLateFilingPenalty(false, 150, DCAAccountStatus, DCADunningStatus)},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is DCA",
				args: args{penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is DCA",
				args: args{penalty: createLateFilingPenalty(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is DCA",
				args: args{penalty: createLateFilingPenalty(false, 150, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is IPEN1",
				args: args{penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is IPEN2",
				args: args{penalty: createLateFilingPenalty(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is CAN",
				args: args{penalty: createLateFilingPenalty(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("CAN"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

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
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

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
				args: args{penalty: createSanctionsPenalty(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount less than 0 and is paid",
				args: args{penalty: createSanctionsPenalty(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: createSanctionsPenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: createRoePenalty(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount less than 0 and is paid",
				args: args{penalty: createRoePenalty(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: createRoePenalty(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is ipen1",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is ipen2",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is ipen3",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN3"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed pending allocation payable status for penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
		}{
			{
				name: "Late filing penalty paid today",
				args: args{penalty: createLateFilingPenalty(true, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions penalty paid today",
				args: args{penalty: createSanctionsPenalty(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions ROE penalty paid today",
				args: args{penalty: createRoePenalty(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &now
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, ClosedPendingAllocationPayableStatus)
			})
		}
	})

	Convey("Get closed payable status for an existing paid penalty from E5 in the same account as a newly paid penalty", t, func() {
		oldPaidPenalty := createRoePenalty(true, 0, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))
		newPaidPenalty := createRoePenalty(true, 250, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))

		closedAt := time.Now()
		e5Transactions := []models.AccountPenaltiesDataDao{*oldPaidPenalty, *newPaidPenalty}

		So(getPayableStatus(types.Penalty.String(), oldPaidPenalty, &closedAt, e5Transactions, allowedTransactionMap, &cfg), ShouldEqual, ClosedPayableStatus)
		So(getPayableStatus(types.Penalty.String(), newPaidPenalty, &closedAt, e5Transactions, allowedTransactionMap, &cfg), ShouldEqual, ClosedPendingAllocationPayableStatus)

	})

	Convey("Get disabled payable status for sanctions - Confirmation Statement", t, func() {
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
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: createSanctionsPenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.FeatureFlagSanctionsCSDisabled = true
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get disabled payable status for sanctions - ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: createRoePenalty(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.FeatureFlagSanctionsROEDisabled = true
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func createLateFilingPenalty(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	return &models.AccountPenaltiesDataDao{
		CompanyCode:          utils.LateFilingPenaltyCompanyCode,
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: "A1234567",
		TransactionDate:      "2025-02-25",
		MadeUpDate:           "2025-02-12",
		Amount:               150,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      "1",
		TransactionSubType:   "EH",
		TypeDescription:      "Penalty Ltd Wel & Eng <=1m    LTDWA",
		DueDate:              "2025-03-26",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	}
}

func createSanctionsPenalty(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	return &models.AccountPenaltiesDataDao{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "E1",
		CustomerCode:         "12345678",
		TransactionReference: "P1234567",
		TransactionDate:      "2025-02-25",
		MadeUpDate:           "2025-02-12",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      SanctionsTransactionType,
		TransactionSubType:   SanctionsTransactionSubType,
		TypeDescription:      "CS01                                    ",
		DueDate:              "2025-03-26",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	}
}

func createRoePenalty(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	return &models.AccountPenaltiesDataDao{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "FU",
		CustomerCode:         "OE123456",
		TransactionReference: "U1234567",
		TransactionDate:      "2025-02-25",
		MadeUpDate:           "2025-02-12",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      SanctionsTransactionType,
		TransactionSubType:   SanctionsRoeFailureToUpdateTransactionSubType,
		TypeDescription:      "PENU                                    ",
		DueDate:              "2025-03-26",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	}
}
