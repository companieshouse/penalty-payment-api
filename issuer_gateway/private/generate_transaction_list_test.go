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

var cfg = config.Config{}
var SanctionsMultipleTransactionSubType = "S1,A2"

func TestUnitGenerateTransactionListFromE5Response(t *testing.T) {
	etag := "ABCDE"
	customerCode := "12345678"
	overSeasEntityId := "OE123456"
	otherTransactionSubType := "Other"
	penaltyTransactionType := "1"
	euTransactionSubType := "EU"
	pen1DunningStatus := addTrailingSpacesToDunningStatus(PEN1DunningStatus)
	dcaDunningStatus := addTrailingSpacesToDunningStatus(DCADunningStatus)
	lfpAccountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
		customerCode, utils.LateFilingPenaltyCompanyCode, euTransactionSubType, pen1DunningStatus, utils.LateFilingPenaltyRefType, false)
	lfpPenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.LateFilingPenaltyRefType)
	sanctionsPenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.SanctionsPenaltyRefType)
	sanctionsRoePenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.SanctionsRoePenaltyRefType)
	transactionAllowed = func(transactionType string, transactionSubtype string) bool {
		return (transactionType == penaltyTransactionType) && (transactionSubtype != otherTransactionSubType)
	}

	Convey("error when first etag generator fails", t, func() {
		etagGenerator = func() (string, error) {
			return "", errors.New("error generating etag")
		}
		penaltyRefType := utils.LateFilingPenaltyRefType

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")

		So(err.Error(), ShouldStartWith, "error generating etag")
		So(transactionList, ShouldBeNil)
	})

	Convey("error when first etag generator succeeds but second etag generator fails", t, func() {
		callCount := 0
		etagGenerator = func() (string, error) {
			callCount++
			if callCount == 2 {
				return "", errors.New("error generating etag")
			}
			return etag, nil
		}
		penaltyRefType := utils.LateFilingPenaltyRefType

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")

		So(err.Error(), ShouldStartWith, "error generating etag")
		So(transactionList, ShouldBeNil)
	})

	Convey("penalty list successfully generated from E5 response - unpaid costs", t, func() {
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.LateFilingPenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			customerCode, utils.LateFilingPenaltyCompanyCode, euTransactionSubType, pen1DunningStatus, penaltyRefType, true)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")

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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.LateFilingPenaltyRefType

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")

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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.LateFilingPenaltyRefType
		otherAccountPenalties := buildTestUnpaidAccountPenaltiesDao(
			customerCode, utils.LateFilingPenaltyCompanyCode, otherTransactionSubType, pen1DunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			otherAccountPenalties, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")

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
		penaltyRefType := utils.LateFilingPenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			customerCode, utils.LateFilingPenaltyCompanyCode, euTransactionSubType, dcaDunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, &cfg, "")
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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.SanctionsPenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			customerCode, utils.SanctionsCompanyCode, SanctionsTransactionSubType, pen1DunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsPenaltyDetailsMap, &cfg, "")

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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.SanctionsRoePenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			overSeasEntityId, utils.SanctionsCompanyCode, SanctionsRoeFailureToUpdateTransactionSubType, pen1DunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsRoePenaltyDetailsMap, &cfg, "")

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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.SanctionsPenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			customerCode, utils.SanctionsCompanyCode, SanctionsTransactionSubType, dcaDunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsPenaltyDetailsMap, &cfg, "")

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
		etagGenerator = func() (string, error) {
			return etag, nil
		}
		penaltyRefType := utils.SanctionsRoePenaltyRefType
		accountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
			overSeasEntityId, utils.SanctionsCompanyCode, SanctionsRoeFailureToUpdateTransactionSubType, dcaDunningStatus, penaltyRefType, false)

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsRoePenaltyDetailsMap, &cfg, "")

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
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount and not paid",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN1",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN2",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN3",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN1",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, DCADunningStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN2",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN3",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN1",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN2",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN3",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN1",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN2",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN3",
				args: args{
					penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount less than 0 and is paid",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(true, -150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA (no trailing spaces)",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, DCAAccountStatus, DCADunningStatus)},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is DCA",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is DCA",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is DCA",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is IPEN1",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is IPEN2",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is CAN",
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("CAN"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount less than 0 and is paid",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount less than 0 and is paid",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is ipen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is ipen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is ipen3",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN3"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestLFPAccountPenaltiesDataDao(true, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions penalty paid today",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions ROE penalty paid today",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &now
				got := getPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

				So(got, ShouldEqual, ClosedPendingAllocationPayableStatus)
			})
		}
	})

	Convey("Get closed payable status for an existing paid penalty from E5 in the same account as a newly paid penalty", t, func() {
		oldPaidPenalty := buildTestROEAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))
		newPaidPenalty := buildTestROEAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))

		closedAt := time.Now()
		e5Transactions := []models.AccountPenaltiesDataDao{*oldPaidPenalty, *newPaidPenalty}

		So(getPayableStatus(types.Penalty.String(), oldPaidPenalty, &closedAt, e5Transactions, &cfg), ShouldEqual, ClosedPayableStatus)
		So(getPayableStatus(types.Penalty.String(), newPaidPenalty, &closedAt, e5Transactions, &cfg), ShouldEqual, ClosedPendingAllocationPayableStatus)

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
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

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
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsRoeFailureToUpdateTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get disabled payable status for sanctions - Multiple Subtypes", t, func() {
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
				args: args{penalty: buildTestSanctionsAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: buildTestROEAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsMultipleTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				got := getPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func buildTestLFPAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.LateFilingPenaltyCompanyCode,
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: "A1234567",
		Amount:               150,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      "1",
		TransactionSubType:   "EU",
		TypeDescription:      "Penalty Ltd Wel & Eng <=1m    LTDWA",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildTestSanctionsAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "E1",
		CustomerCode:         "12345678",
		TransactionReference: "P1234567",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      SanctionsTransactionType,
		TransactionSubType:   SanctionsTransactionSubType,
		TypeDescription:      "CS01                                    ",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildTestROEAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "FU",
		CustomerCode:         "OE123456",
		TransactionReference: "U1234567",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      SanctionsTransactionType,
		TransactionSubType:   SanctionsRoeFailureToUpdateTransactionSubType,
		TypeDescription:      "PENU                                    ",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

type AccountPenaltiesParams struct {
	CompanyCode          string
	LedgerCode           string
	CustomerCode         string
	TransactionReference string
	Amount               float64
	OutstandingAmount    float64
	IsPaid               bool
	TransactionType      string
	TransactionSubType   string
	TypeDescription      string
	AccountStatus        string
	DunningStatus        string
}

func buildTestAccountPenaltiesDataDao(params AccountPenaltiesParams) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          params.CompanyCode,
		LedgerCode:           params.LedgerCode,
		CustomerCode:         params.CustomerCode,
		TransactionReference: params.TransactionReference,
		TransactionDate:      "2025-02-25",
		MadeUpDate:           "2025-02-12",
		Amount:               params.Amount,
		OutstandingAmount:    params.OutstandingAmount,
		IsPaid:               params.IsPaid,
		TransactionType:      params.TransactionType,
		TransactionSubType:   params.TransactionSubType,
		TypeDescription:      params.TypeDescription,
		DueDate:              "2025-03-26",
		AccountStatus:        params.AccountStatus,
		DunningStatus:        params.DunningStatus,
	}
}

func buildTestUnpaidAccountPenaltiesDao(customerCode, companyCode, transactionSubType, dunningStatus, penaltyRefType string, withUnpaidCost bool) *models.AccountPenaltiesDao {
	params := AccountPenaltiesParams{
		CompanyCode:        companyCode,
		CustomerCode:       customerCode,
		TransactionSubType: transactionSubType,
		Amount:             250,
		OutstandingAmount:  250,
		IsPaid:             false,
		AccountStatus:      CHSAccountStatus,
		DunningStatus:      dunningStatus,
	}
	switch penaltyRefType {
	case utils.SanctionsPenaltyRefType:
		{
			params.TransactionReference = "P1234567"
			params.TransactionType = SanctionsTransactionType
			params.LedgerCode = "E1"
			params.TypeDescription = "CS01                                    "
		}
	case utils.LateFilingPenaltyRefType:
		{
			params.TransactionReference = "A1234567"
			params.TransactionType = "1"
			params.LedgerCode = "EW"
			params.TypeDescription = "Penalty Ltd Wel & Eng <=1m    LTDWA"
		}
	default:
		{
			params.TransactionReference = "U1234567"
			params.TransactionType = SanctionsTransactionType
			params.LedgerCode = "FU"
			params.TypeDescription = "Penalty Ltd Wel & Eng <=1m    LTDWA"
		}
	}
	accountPenaltiesDataDao := []models.AccountPenaltiesDataDao{
		buildTestAccountPenaltiesDataDao(params),
	}
	if withUnpaidCost {
		params.TransactionReference = "F1"
		params.TransactionType = "2"
		accountPenaltiesDataDao = append(accountPenaltiesDataDao, buildTestAccountPenaltiesDataDao(params))
	}
	return &models.AccountPenaltiesDao{
		CustomerCode:     customerCode,
		CompanyCode:      companyCode,
		CreatedAt:        &now,
		AccountPenalties: accountPenaltiesDataDao,
	}
}

func buildTestPenaltyDetailsMap(penaltyRefType string) *config.PenaltyDetailsMap {
	penaltyDetailsMap := config.PenaltyDetailsMap{
		Name:    "penalty details",
		Details: map[string]config.PenaltyDetails{},
	}
	switch penaltyRefType {
	case utils.LateFilingPenaltyRefType:
		penaltyDetailsMap.Details[penaltyRefType] = config.PenaltyDetails{
			Description:        "Late Filing Penalty",
			DescriptionId:      "late-filing-penalty",
			ClassOfPayment:     "penalty-lfp",
			ResourceKind:       "late-filing-penalty#late-filing-penalty",
			ProductType:        "late-filing-penalty",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		}
	case utils.SanctionsPenaltyRefType:
		penaltyDetailsMap.Details[penaltyRefType] = config.PenaltyDetails{
			Description:        "Sanctions Penalty Payment",
			DescriptionId:      "penalty-sanctions",
			ClassOfPayment:     "penalty-sanctions",
			ResourceKind:       "penalty#sanctions",
			ProductType:        "penalty-sanctions",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		}
	case utils.SanctionsRoePenaltyRefType:
		penaltyDetailsMap.Details[penaltyRefType] = config.PenaltyDetails{
			Description:        "Overseas Entity Penalty Payment",
			DescriptionId:      "penalty-sanctions",
			ClassOfPayment:     "penalty-sanctions",
			ResourceKind:       "penalty#sanctions",
			ProductType:        "penalty-sanctions",
			EmailReceivedAppId: "penalty-payment-api.sanctions_roe_penalty_payment_received_email",
			EmailMsgType:       "sanctions_roe_penalty_payment_received_email",
		}
	}

	return &penaltyDetailsMap
}
