package private

import (
	"fmt"
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var etagGenerator = utils.GenerateEtag

func GenerateTransactionListFromAccountPenalties(accountPenalties *models.AccountPenaltiesDao, companyCode string,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, error) {
	payableTransactionList := models.TransactionListResponse{}
	etag, err := etagGenerator()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.Error(err)
		return nil, err
	}

	payableTransactionList.Etag = etag
	payableTransactionList.TotalResults = len(accountPenalties.AccountPenalties)

	// Loop through penalties and construct CH resources
	for _, accountPenalty := range accountPenalties.AccountPenalties {
		transactionListItem, err := buildTransactionListItemFromAccountPenalty(
			&accountPenalty, allowedTransactionsMap, penaltyDetailsMap, companyCode, accountPenalties.ClosedAt,
			accountPenalties.AccountPenalties)
		if err != nil {
			return nil, err
		}

		payableTransactionList.Items = append(payableTransactionList.Items, transactionListItem)
	}

	return &payableTransactionList, nil
}

func buildTransactionListItemFromAccountPenalty(e5Transaction *models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap,
	penaltyDetailsMap *config.PenaltyDetailsMap, companyCode string, closedAt *time.Time,
	e5Transactions []models.AccountPenaltiesDataDao) (models.TransactionListItem, error) {
	etag, err := utils.GenerateEtag()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.Error(err)
		return models.TransactionListItem{}, err
	}

	transactionType := getTransactionType(e5Transaction, allowedTransactionsMap)

	transactionListItem := models.TransactionListItem{}
	transactionListItem.Etag = etag
	transactionListItem.ID = e5Transaction.TransactionReference
	transactionListItem.IsPaid = e5Transaction.IsPaid
	transactionListItem.Kind = penaltyDetailsMap.Details[companyCode].ResourceKind
	transactionListItem.IsDCA = checkDunningStatus(e5Transaction, DCADunningStatus)
	transactionListItem.DueDate = e5Transaction.DueDate
	transactionListItem.MadeUpDate = e5Transaction.MadeUpDate
	transactionListItem.TransactionDate = e5Transaction.TransactionDate
	transactionListItem.OriginalAmount = e5Transaction.Amount
	transactionListItem.Outstanding = e5Transaction.OutstandingAmount
	transactionListItem.Type = transactionType
	transactionListItem.Reason = getReason(e5Transaction)
	transactionListItem.PayableStatus = getPayableStatus(transactionType, e5Transaction, closedAt, e5Transactions, allowedTransactionsMap)

	return transactionListItem, nil
}

func getTransactionType(e5Transaction *models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap) string {
	// Each penalty needs to be checked and identified as a 'penalty' or 'other'. This allows penalty-payment-web to determine
	// which transactions are payable. This is done using a yaml file to map payable transactions
	// Check if the penalty is allowed and set to 'penalty' if it is
	if _, ok := allowedTransactionsMap.Types[e5Transaction.TransactionType][e5Transaction.TransactionSubType]; ok {
		return types.Penalty.String()
	} else {
		return types.Other.String()
	}
}

func getReason(transaction *models.AccountPenaltiesDataDao) string {
	switch transaction.CompanyCode {
	case utils.LateFilingPenaltyCompanyCode:
		return LateFilingPenaltyReason
	case utils.SanctionsCompanyCode:
		return getSanctionsReason(transaction)
	default:
		return PenaltyReason
	}
}

func getSanctionsReason(transaction *models.AccountPenaltiesDataDao) string {
	if transaction.TransactionSubType == SanctionsTransactionSubType &&
		strings.TrimSpace(transaction.TypeDescription) == CS01TypeDescription {
		return ConfirmationStatementReason
	} else if transaction.TransactionSubType == SanctionsRoeFailureToUpdateTransactionSubType {
		return SanctionsRoeFailureToUpdateReason
	} else {
		return PenaltyReason
	}
}

const (
	SanctionsTransactionType                      = "1"
	SanctionsTransactionSubType                   = "S1"
	SanctionsRoeFailureToUpdateTransactionSubType = "A2"
	CS01TypeDescription                           = "CS01"

	LateFilingPenaltyReason           = "Late filing of accounts"
	ConfirmationStatementReason       = "Failure to file a confirmation statement"
	SanctionsRoeFailureToUpdateReason = "Failure to update the Register of Overseas Entities"
	PenaltyReason                     = "Penalty"

	OpenPayableStatus                    = "OPEN"
	ClosedPayableStatus                  = "CLOSED"
	ClosedPendingAllocationPayableStatus = "CLOSED_PENDING_ALLOCATION"

	CHSAccountStatus = "CHS"
	DCAAccountStatus = "DCA"
	HLDAccountStatus = "HLD"
	WDRAccountStatus = "WDR"

	DCADunningStatus  = "DCA"
	PEN1DunningStatus = "PEN1"
	PEN2DunningStatus = "PEN2"
	PEN3DunningStatus = "PEN3"
)

func getPayableStatus(transactionType string, e5Transaction *models.AccountPenaltiesDataDao, closedAt *time.Time,
	e5Transactions []models.AccountPenaltiesDataDao, allowedTransactionsMap *models.AllowedTransactionMap) string {
	if types.Penalty.String() == transactionType {
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
		penaltyPaidToday(closedAt) {
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
