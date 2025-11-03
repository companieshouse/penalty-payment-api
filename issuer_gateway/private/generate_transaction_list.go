package private

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var etagGenerator = utils.GenerateEtag

func GenerateTransactionListFromAccountPenalties(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap, cfg *config.Config, requestId string,
	reasonProvider ReasonProvider, payableStatusProvider PayableStatusProvider) (*models.TransactionListResponse, error) {
	payableTransactionList := models.TransactionListResponse{}
	etag, err := etagGenerator()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.ErrorC(requestId, err)
		return nil, err
	}

	payableTransactionList.Etag = etag
	payableTransactionList.TotalResults = len(accountPenalties.AccountPenalties)

	// Loop through penalties and construct CH resources
	for _, accountPenalty := range accountPenalties.AccountPenalties {
		transactionType := getTransactionType(&accountPenalty, allowedTransactionsMap)
		reason := reasonProvider.GetReason(&accountPenalty)
		payableStatus := payableStatusProvider.GetPayableStatus(transactionType, &accountPenalty, accountPenalties.ClosedAt, accountPenalties.AccountPenalties, allowedTransactionsMap, cfg)
		transactionListItem, err := buildTransactionListItemFromAccountPenalty(&accountPenalty, penaltyDetailsMap, penaltyRefType, transactionType, reason, payableStatus, requestId)
		if err != nil {
			return nil, err
		}

		payableTransactionList.Items = append(payableTransactionList.Items, transactionListItem)
	}

	return &payableTransactionList, nil
}

func buildTransactionListItemFromAccountPenalty(dao *models.AccountPenaltiesDataDao,
	penaltyDetailsMap *config.PenaltyDetailsMap, penaltyRefType string, transactionType string,
	reason string, payableStatus string, requestId string) (models.TransactionListItem, error) {
	etag, err := etagGenerator()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.ErrorC(requestId, err)
		return models.TransactionListItem{}, err
	}

	transactionListItem := models.TransactionListItem{}
	transactionListItem.Etag = etag
	transactionListItem.ID = dao.TransactionReference
	transactionListItem.IsPaid = dao.IsPaid
	transactionListItem.Kind = penaltyDetailsMap.Details[penaltyRefType].ResourceKind
	transactionListItem.IsDCA = checkDunningStatus(dao, DCADunningStatus)
	transactionListItem.DueDate = dao.DueDate
	transactionListItem.MadeUpDate = dao.MadeUpDate
	transactionListItem.TransactionDate = dao.TransactionDate
	transactionListItem.OriginalAmount = dao.Amount
	transactionListItem.Outstanding = dao.OutstandingAmount
	transactionListItem.Type = transactionType
	transactionListItem.Reason = reason
	transactionListItem.PayableStatus = payableStatus

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

const (
	InvoiceTransactionType                            = "1"
	SanctionsConfirmationStatementTransactionSubType  = "S1"
	SanctionsFailedToVerifyIdentityTransactionSubType = "S3"
	SanctionsRoeFailureToUpdateTransactionSubType     = "A2"

	CHSAccountStatus = "CHS"
	DCAAccountStatus = "DCA"
	HLDAccountStatus = "HLD"
	WDRAccountStatus = "WDR"

	DCADunningStatus  = "DCA"
	PEN1DunningStatus = "PEN1"
	PEN2DunningStatus = "PEN2"
	PEN3DunningStatus = "PEN3"
)
