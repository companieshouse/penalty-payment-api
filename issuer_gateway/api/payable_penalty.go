package api

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var getAccountPenalties = AccountPenalties
var getMatchingPenalty = private.MatchPenalty

func PayablePenalty(penaltyRefType, customerCode, companyCode string, transaction models.TransactionItem, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionItem, error) {

	response, _, err := getAccountPenalties(penaltyRefType, customerCode, companyCode, penaltyDetailsMap, allowedTransactionsMap, apDaoSvc)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	unpaidPenaltyCount := getUnpaidPenaltyCount(response.Items)
	if unpaidPenaltyCount > 1 {
		log.Info("customer has more than one outstanding penalty", log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
			"penalty_count": unpaidPenaltyCount,
		})
	}

	return getMatchingPenalty(response.Items, transaction, customerCode)
}

func getUnpaidPenaltyCount(transactionListItems []models.TransactionListItem) int {
	unpaidPenaltyCount := 0
	for _, tx := range transactionListItems {
		if !tx.IsPaid && (tx.Type == types.Penalty.String()) {
			unpaidPenaltyCount++
		}
	}

	return unpaidPenaltyCount
}
