package api

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var getAccountPenalties = AccountPenalties
var getMatchingPenalty = private.MatchPenalty

func PayablePenalty(companyNumber string, companyCode string, transaction models.TransactionItem,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionItem, error) {

	response, _, err := getAccountPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// for the first release, the company must only have one outstanding penalty
	unpaidPenaltyCount := getUnpaidPenaltyCount(response.Items)
	if unpaidPenaltyCount > 1 {
		log.Info("company has more than one outstanding penalty", log.Data{
			"company_number": companyNumber,
			"penalty_count":  unpaidPenaltyCount,
		})
		return nil, private.ErrMultiplePenalties
	}

	return getMatchingPenalty(response.Items, transaction, companyNumber)
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
