package types

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/config"
)

type AccountPenaltiesParams struct {
	PenaltyRefType             string
	CustomerCode               string
	CompanyCode                string
	PenaltyDetailsMap          *config.PenaltyDetailsMap
	AllowedTransactionsMap     *models.AllowedTransactionMap
	AccountPenaltiesDaoService dao.AccountPenaltiesDaoService
	RequestId                  string
}

type PayablePenaltyParams struct {
	PenaltyRefType             string
	CustomerCode               string
	CompanyCode                string
	Transaction                models.TransactionItem
	PenaltyDetailsMap          *config.PenaltyDetailsMap
	AllowedTransactionsMap     *models.AllowedTransactionMap
	AccountPenaltiesDaoService dao.AccountPenaltiesDaoService
	RequestId                  string
}
