package private

import (
	"fmt"
	"strings"

	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var etagGenerator = utils.GenerateEtag

func GenerateTransactionListFromE5Response(e5Response *e5.GetTransactionsResponse, companyCode string,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, error) {
	payableTransactionList := models.TransactionListResponse{}
	etag, err := etagGenerator()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.Error(err)
		return nil, err
	}

	payableTransactionList.Etag = etag
	payableTransactionList.TotalResults = e5Response.Page.TotalElements

	// Loop through e5 response and construct CH resources
	for _, e5Transaction := range e5Response.Transactions {
		transactionListItem, err := buildTransactionListItemFromE5Transaction(
			&e5Transaction, allowedTransactionsMap, penaltyDetailsMap, companyCode)
		if err != nil {
			return nil, err
		}

		payableTransactionList.Items = append(payableTransactionList.Items, transactionListItem)
	}

	return &payableTransactionList, nil
}

func buildTransactionListItemFromE5Transaction(e5Transaction *e5.Transaction, allowedTransactionsMap *models.AllowedTransactionMap, penaltyDetailsMap *config.PenaltyDetailsMap, companyCode string) (models.TransactionListItem, error) {
	etag, err := utils.GenerateEtag()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.Error(err)
		return models.TransactionListItem{}, err
	}

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
	transactionListItem.Type = getTransactionType(e5Transaction, allowedTransactionsMap)
	transactionListItem.Reason = getReason(e5Transaction)
	transactionListItem.PayableStatus = getPayableStatus(e5Transaction)

	return transactionListItem, nil
}

func getTransactionType(e5Transaction *e5.Transaction, allowedTransactionsMap *models.AllowedTransactionMap) string {
	// Each transaction needs to be checked and identified as a 'penalty' or 'other'. This allows penalty-payment-web to determine
	// which transactions are payable. This is done using a yaml file to map payable transactions
	// Check if the transaction is allowed and set to 'penalty' if it is
	if _, ok := allowedTransactionsMap.Types[e5Transaction.TransactionType][e5Transaction.TransactionSubType]; ok {
		return types.Penalty.String()
	} else {
		return types.Other.String()
	}
}

func getReason(transaction *e5.Transaction) string {
	if transaction.CompanyCode == utils.LateFilingPenalty {
		return LateFilingPenaltyReason
	} else if transaction.CompanyCode == utils.Sanctions && checkSanctionsTypeDescription(transaction, CS01TypeDescription) {
		return ConfirmationStatementReason
	}
	return PenaltyReason
}

func checkSanctionsTypeDescription(transaction *e5.Transaction, typeDescription string) bool {
	return (transaction.TransactionType == SanctionsTransactionType && transaction.TransactionSubType == SanctionsTransactionSubType) &&
		strings.TrimSpace(transaction.TypeDescription) == typeDescription
}

const (
	SanctionsTransactionType    = "1"
	SanctionsTransactionSubType = "S1"
	CS01TypeDescription         = "CS01"

	LateFilingPenaltyReason     = "Late filing of accounts"
	ConfirmationStatementReason = "Failure to file a confirmation statement"
	PenaltyReason               = "Penalty"

	OpenPayableStatus   = "OPEN"
	ClosedPayableStatus = "CLOSED"

	CHSAccountStatus = "CHS"
	DCAAccountStatus = "DCA"
	HLDAccountStatus = "HLD"
	WDRAccountStatus = "WDR"

	DCADunningStatus  = "DCA"
	PEN1DunningStatus = "PEN1"
	PEN2DunningStatus = "PEN2"
	PEN3DunningStatus = "PEN3"
)

func getPayableStatus(transaction *e5.Transaction) string {
	if transaction.IsPaid || transaction.OutstandingAmount <= 0 || checkDunningStatus(transaction, DCADunningStatus) {
		return ClosedPayableStatus
	}

	if transaction.CompanyCode == utils.LateFilingPenalty &&
		(checkDunningStatus(transaction, PEN1DunningStatus) || checkDunningStatus(transaction, PEN2DunningStatus) || checkDunningStatus(transaction, PEN3DunningStatus)) &&
		(transaction.AccountStatus == CHSAccountStatus || transaction.AccountStatus == DCAAccountStatus || transaction.AccountStatus == HLDAccountStatus || transaction.AccountStatus == WDRAccountStatus) {
		return OpenPayableStatus
	} else if transaction.CompanyCode == utils.Sanctions &&
		(checkDunningStatus(transaction, PEN1DunningStatus) || checkDunningStatus(transaction, PEN2DunningStatus)) &&
		(transaction.AccountStatus == CHSAccountStatus || transaction.AccountStatus == DCAAccountStatus || transaction.AccountStatus == HLDAccountStatus) {
		return OpenPayableStatus
	}

	return ClosedPayableStatus
}

func checkDunningStatus(transaction *e5.Transaction, dunningStatus string) bool {
	return strings.TrimSpace(transaction.DunningStatus) == dunningStatus
}
