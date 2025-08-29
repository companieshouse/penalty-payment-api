package api

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var getAccountPenalties = AccountPenalties
var getMatchingPenalty = private.MatchPenalty

func PayablePenalty(params types.PayablePenaltyParams) (*models.TransactionItem, error) {
	penaltyRefType := params.PenaltyRefType
	customerCode := params.CustomerCode
	companyCode := params.CompanyCode
	transaction := params.Transaction
	penaltyDetailsMap := params.PenaltyDetailsMap
	apDaoSvc := params.AccountPenaltiesDaoService
	allowedTransactionsMap := params.AllowedTransactionsMap
	context := params.Context

	accountPenaltiesParams := types.AccountPenaltiesParams{
		PenaltyRefType:             penaltyRefType,
		CustomerCode:               customerCode,
		CompanyCode:                companyCode,
		PenaltyDetailsMap:          penaltyDetailsMap,
		AllowedTransactionsMap:     allowedTransactionsMap,
		AccountPenaltiesDaoService: apDaoSvc,
		Context:                    context,
	}
	response, _, err := getAccountPenalties(accountPenaltiesParams)
	if err != nil {
		log.ErrorC(context, err)
		return nil, err
	}

	unpaidPenaltyCount := getUnpaidPenaltyCount(response.Items)
	log.InfoC(context, "unpaid penalties", log.Data{
		"unpaid_penalties_count": unpaidPenaltyCount,
		"customer_code":          customerCode,
		"company_code":           companyCode,
	})

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
