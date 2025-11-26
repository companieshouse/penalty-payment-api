package private

import (
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

const (
	OpenPayableStatus                       = "OPEN"
	DisabledPayableStatus                   = "DISABLED"
	ClosedPayableStatus                     = "CLOSED"
	ClosedPendingAllocationPayableStatus    = "CLOSED_PENDING_ALLOCATION"
	ClosedInstalmentPlanPayableStatus       = "CLOSED_INSTALMENT_PLAN"
	ClosedPenStrategyExhaustedPayableStatus = "CLOSED_PEN_STRATEGY_EXHAUSTED"
)

type PayableStatusProvider interface {
	GetPayableStatus(transactionType string, e5Transaction *models.AccountPenaltiesDataDao, closedAt *time.Time,
		e5Transactions []models.AccountPenaltiesDataDao, penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig,
		cfg *config.Config, financePaymentConfig finance_config.FinancePaymentConfig) string
}

type DefaultPayableStatusProvider struct{}

func (provider *DefaultPayableStatusProvider) GetPayableStatus(transactionType string, e5Transaction *models.AccountPenaltiesDataDao,
	closedAt *time.Time, e5Transactions []models.AccountPenaltiesDataDao, penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig,
	cfg *config.Config, financePaymentConfig finance_config.FinancePaymentConfig) string {
	if types.Penalty.String() == transactionType {
		txType := e5Transaction.TransactionType
		txSubtype := e5Transaction.TransactionSubType

		if penaltyConfigTransactionSubTypeDisabled(txSubtype, cfg) {
			return DisabledPayableStatus
		}

		found, disabled := penaltyTypeTransactionSubTypeStatus(penaltyTypes, txType, txSubtype)
		if found {
			if disabled {
				return DisabledPayableStatus
			}
		} else {
			return ClosedPayableStatus
		}

		if payablePenaltySubTypeAbsent(financePaymentConfig, txType, txSubtype) {
			return ClosedPayableStatus
		}

		closedPayableStatus, isClosed := checkClosedPayableStatus(e5Transaction, closedAt, e5Transactions, penaltyTypes)
		if isClosed {
			return closedPayableStatus
		}

		openPayableStatus, isOpen := checkOpenPayableStatus(e5Transaction)
		if isOpen {
			return openPayableStatus
		}
	}

	return ClosedPayableStatus
}

func checkClosedPayableStatus(penalty *models.AccountPenaltiesDataDao, closedAt *time.Time, e5Transactions []models.AccountPenaltiesDataDao,
	penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig) (payableStatus string, isClosed bool) {
	if (penalty.IsPaid && closedAt != nil) &&
		penaltyPaidToday(closedAt) &&
		!penaltyPaymentAllocated(penalty) {
		return ClosedPendingAllocationPayableStatus, true
	}

	if checkClosedInstalmentPlanPayableStatus(penalty, e5Transactions) {
		return ClosedInstalmentPlanPayableStatus, true
	}

	if checkClosedPenStrategyExhaustedPayableStatus(penalty, e5Transactions) {
		return ClosedPenStrategyExhaustedPayableStatus, true
	}

	if penalty.IsPaid || penalty.OutstandingAmount <= 0 || checkDunningStatus(penalty, DCADunningStatus) ||
		len(getUnpaidCosts(penalty, e5Transactions, penaltyTypes)) > 0 {
		return ClosedPayableStatus, true
	}
	return "", false
}

func checkClosedInstalmentPlanPayableStatus(penalty *models.AccountPenaltiesDataDao, e5Transactions []models.AccountPenaltiesDataDao) bool {
	for _, e5Transaction := range e5Transactions {
		if isInstalmentPlanTransaction(penalty, e5Transaction) {
			return true
		}
	}
	return false
}

func isInstalmentPlanTransaction(penalty *models.AccountPenaltiesDataDao, e5Transaction models.AccountPenaltiesDataDao) bool {
	return e5Transaction.MadeUpDate == penalty.MadeUpDate &&
		(e5Transaction.TransactionType == "P" && e5Transaction.TransactionSubType == "00")
}

func checkClosedPenStrategyExhaustedPayableStatus(penalty *models.AccountPenaltiesDataDao, e5Transactions []models.AccountPenaltiesDataDao) bool {
	for _, e5Transaction := range e5Transactions {
		if isExhaustedWriteOffTransaction(penalty, e5Transaction) {
			return true
		}
	}
	return false
}

func isExhaustedWriteOffTransaction(penalty *models.AccountPenaltiesDataDao, e5Transaction models.AccountPenaltiesDataDao) bool {
	return e5Transaction.MadeUpDate == penalty.MadeUpDate &&
		(e5Transaction.TransactionType == "4" && e5Transaction.TransactionSubType == "82")
}

func getUnpaidCosts(penalty *models.AccountPenaltiesDataDao, e5Transactions []models.AccountPenaltiesDataDao,
	penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig) (unpaidCosts []models.AccountPenaltiesDataDao) {
	for _, e5Transaction := range e5Transactions {
		transactionType := getTransactionType(&e5Transaction, penaltyTypes)
		if (e5Transaction.TransactionReference != penalty.TransactionReference && !e5Transaction.IsPaid) &&
			(types.Other.String() == transactionType && penalty.MadeUpDate == e5Transaction.MadeUpDate) {
			unpaidCosts = append(unpaidCosts, e5Transaction)
		}
	}
	return unpaidCosts
}

func checkOpenPayableStatus(penalty *models.AccountPenaltiesDataDao) (payableStatus string, isOpen bool) {
	if penalty.CompanyCode == utils.LateFilingPenaltyCompanyCode &&
		(checkDunningStatus(penalty, PEN1DunningStatus) || checkDunningStatus(penalty, PEN2DunningStatus) || checkDunningStatus(penalty, PEN3DunningStatus)) &&
		(penalty.AccountStatus == CHSAccountStatus || penalty.AccountStatus == DCAAccountStatus || penalty.AccountStatus == HLDAccountStatus || penalty.AccountStatus == WDRAccountStatus) {
		return OpenPayableStatus, true
	} else if penalty.CompanyCode == utils.SanctionsCompanyCode &&
		(checkDunningStatus(penalty, PEN1DunningStatus) || checkDunningStatus(penalty, PEN2DunningStatus)) &&
		(penalty.AccountStatus == CHSAccountStatus || penalty.AccountStatus == DCAAccountStatus || penalty.AccountStatus == HLDAccountStatus) {
		return OpenPayableStatus, true
	}
	return "", false
}

func penaltyPaidToday(closedAt *time.Time) bool {
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := closedAt.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func checkDunningStatus(transaction *models.AccountPenaltiesDataDao, dunningStatus string) bool {
	return strings.TrimSpace(transaction.DunningStatus) == dunningStatus
}

func penaltyPaymentAllocated(penalty *models.AccountPenaltiesDataDao) bool {
	// The value of outstanding amount is 0 after penalty payment is allocated in E5
	// and AccountPenalties cache is updated with E5 data
	return penalty.OutstandingAmount == 0
}

// penaltyConfigTransactionSubTypeDisabled checks the config override that is set via the DISABLED_PENALTY_TRANSACTION_SUBTYPES
// environment variable that can be used to disable a subtype
func penaltyConfigTransactionSubTypeDisabled(txSubtype string, cfg *config.Config) bool {
	log.Debug("Disabled subtypes: ", log.Data{
		"DISABLED_PENALTY_TRANSACTION_SUBTYPES": cfg.DisabledPenaltyTransactionSubtypes})
	trimDisabledSubtypes := strings.ReplaceAll(cfg.DisabledPenaltyTransactionSubtypes, " ", "")
	disabledSubtypes := strings.Split(trimDisabledSubtypes, ",")
	penaltySubType := txSubtype
	for _, subType := range disabledSubtypes {
		if penaltySubType == subType {
			return true
		}
	}

	return false
}

// penaltyTypeTransactionSubTypeStatus returns a boolean to indicate if the subtype has been found in the financial
// penalty type config, and a boolean to return the status of the disabled config value if set
func penaltyTypeTransactionSubTypeStatus(penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig, txType, txSubType string) (bool, bool) {
	subtypesMap, ok := penaltyTypes[txType]
	if !ok || subtypesMap == nil {
		log.Info("Penalty Type Config transaction type not found: ", log.Data{
			"transaction_type": txType})
		return false, false
	}

	subtypeCfg, ok := subtypesMap[txSubType]
	if !ok {
		log.Info("Penalty Type Config transaction subtype not found: ", log.Data{
			"transaction_subtype": txSubType})
		return false, false
	}

	log.Info("Penalty Type Config transaction subtype not found: ", log.Data{
		"transaction_subtype": txSubType,
		"disabled":            subtypeCfg.Disabled})

	return true, subtypeCfg.Disabled
}

// payablePenaltySubTypeAbsent returns a boolean to indicate if the subtype exists in the finance payment configuration
func payablePenaltySubTypeAbsent(financePaymentConfig finance_config.FinancePaymentConfig, txType, txSubType string) bool {
	if financePaymentConfig.TransactionType != txType {
		log.Info("Finance Payment Config transaction subtype not found: ", log.Data{
			"transaction_subtype": txType})
		return true
	}

	for _, v := range financePaymentConfig.TransactionSubTypes {
		if v == txSubType {
			return false
		}
	}

	log.Info("Finance Payment Config transaction subtype not found: ", log.Data{
		"transaction_subtype": txSubType})

	return true
}
