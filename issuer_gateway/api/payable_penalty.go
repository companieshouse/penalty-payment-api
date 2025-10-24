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
	requestId := params.RequestId

	accountPenaltiesParams := types.AccountPenaltiesParams{
		PenaltyRefType:             penaltyRefType,
		CustomerCode:               customerCode,
		CompanyCode:                companyCode,
		PenaltyDetailsMap:          penaltyDetailsMap,
		AccountPenaltiesDaoService: apDaoSvc,
		RequestId:                  requestId,
	}
	response, _, err := getAccountPenalties(accountPenaltiesParams)
	if err != nil {
		log.ErrorC(requestId, err)
		return nil, err
	}

	unpaidPenaltyCount := getUnpaidPenaltyCount(response.Items)
	log.InfoC(requestId, "unpaid penalties", log.Data{
		"unpaid_penalties_count": unpaidPenaltyCount,
		"customer_code":          customerCode,
		"company_code":           companyCode,
	})

	return getMatchingPenalty(response.Items, transaction, customerCode, requestId)
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
