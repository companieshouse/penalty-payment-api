package private

import (
	"strings"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

const (
	OpenPayableStatus                    = "OPEN"
	DisabledPayableStatus                = "DISABLED"
	ClosedPayableStatus                  = "CLOSED"
	ClosedPendingAllocationPayableStatus = "CLOSED_PENDING_ALLOCATION"
)

type PayableStatusProvider interface {
	GetPayableStatus(transactionType string, e5Transaction *models.AccountPenaltiesDataDao, closedAt *time.Time,
		e5Transactions []models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap, cfg *config.Config) string
}

type DefaultPayableStatusProvider struct{}

func (provider *DefaultPayableStatusProvider) GetPayableStatus(transactionType string, e5Transaction *models.AccountPenaltiesDataDao, closedAt *time.Time,
	e5Transactions []models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap, cfg *config.Config) string {
	if types.Penalty.String() == transactionType {
		if penaltyTransactionSubTypeDisabled(e5Transaction, cfg) {
			return DisabledPayableStatus
		}
		closedPayableStatus, isClosed := checkClosedPayableStatus(e5Transaction, closedAt, e5Transactions, allowedTransactionsMap)
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

func checkClosedPayableStatus(penalty *models.AccountPenaltiesDataDao, closedAt *time.Time,
	e5Transactions []models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap) (payableStatus string, isClosed bool) {
	if (penalty.IsPaid && closedAt != nil) &&
		penaltyPaidToday(closedAt) &&
		!penaltyPaymentAllocated(penalty) {
		return ClosedPendingAllocationPayableStatus, true
	}

	if penalty.IsPaid || penalty.OutstandingAmount <= 0 || checkDunningStatus(penalty, DCADunningStatus) ||
		len(getUnpaidCosts(penalty, e5Transactions, allowedTransactionsMap)) > 0 {
		return ClosedPayableStatus, true
	}
	return "", false
}

func getUnpaidCosts(penalty *models.AccountPenaltiesDataDao, e5Transactions []models.AccountPenaltiesDataDao,
	allowedTransactionsMap *models.AllowedTransactionMap) (unpaidCosts []models.AccountPenaltiesDataDao) {
	for _, e5Transaction := range e5Transactions {
		transactionType := getTransactionType(&e5Transaction, allowedTransactionsMap)
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

func penaltyTransactionSubTypeDisabled(penalty *models.AccountPenaltiesDataDao, cfg *config.Config) bool {
	trimDisabledSubtypes := strings.ReplaceAll(cfg.DisabledPenaltyTransactionSubtypes, " ", "")
	disabledSubtypes := strings.Split(trimDisabledSubtypes, ",")
	penaltySubType := penalty.TransactionSubType
	for _, subType := range disabledSubtypes {
		if penaltySubType == subType {
			return true
		}
	}
	return false
}
