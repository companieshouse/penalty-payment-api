package private

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

var now = time.Now().Truncate(time.Millisecond)
var yesterday = time.Now().AddDate(0, 0, -1).Truncate(time.Millisecond)
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

var cfg = config.Config{}
var SanctionsMultipleTransactionSubType = "S1,A2"

func TestUnitGenerateTransactionListFromE5Response(t *testing.T) {
	etag := "ABCDE"
	customerCode := "12345678"
	overSeasEntityId := "OE123456"
	otherTransactionSubType := "Other"
	euTransactionSubType := "EU"
	pen1DunningStatus := addTrailingSpacesToDunningStatus(PEN1DunningStatus)
	dcaDunningStatus := addTrailingSpacesToDunningStatus(DCADunningStatus)
	lfpAccountPenaltiesDao := buildTestUnpaidAccountPenaltiesDao(
		customerCode, utils.LateFilingPenaltyCompanyCode, euTransactionSubType, pen1DunningStatus, utils.LateFilingPenaltyRefType, false)
	lfpPenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.LateFilingPenaltyRefType)
	sanctionsPenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.SanctionsPenaltyRefType)
	sanctionsRoePenaltyDetailsMap := buildTestPenaltyDetailsMap(utils.SanctionsRoePenaltyRefType)

	Convey("error when first etag generator fails", t, func() {
		etagGenerator = func() (string, error) {
			return "", errors.New("error generating etag")
		}
		penaltyRefType := utils.LateFilingPenaltyRefType
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			lfpAccountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			otherAccountPenalties, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, lfpPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)
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
			customerCode, utils.SanctionsCompanyCode, SanctionsConfirmationStatementTransactionSubType, pen1DunningStatus, penaltyRefType, false)
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
			Reason:          SanctionsConfirmationStatementReason,
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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsRoePenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
			customerCode, utils.SanctionsCompanyCode, SanctionsConfirmationStatementTransactionSubType, dcaDunningStatus, penaltyRefType, false)
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsPenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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
			Reason:          SanctionsConfirmationStatementReason,
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
		reasonProvider := &DefaultReasonProvider{}
		payableStatusProvider := &DefaultPayableStatusProvider{}

		transactionList, err := GenerateTransactionListFromAccountPenalties(
			accountPenaltiesDao, penaltyRefType, sanctionsRoePenaltyDetailsMap, allowedTransactionMap, &cfg, "", reasonProvider, payableStatusProvider)

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

func addTrailingSpacesToDunningStatus(dunningStatus string) string {
	return fmt.Sprintf("%-12s", dunningStatus)
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
			params.TransactionType = InvoiceTransactionType
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
			params.TransactionType = InvoiceTransactionType
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
