package types

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
)

type AccountPenaltiesParams struct {
	PenaltyRefType             string
	CustomerCode               string
	CompanyCode                string
	AccountPenaltiesDaoService dao.AccountPenaltiesDaoService
	RequestId                  string
}

type PayablePenaltyParams struct {
	PenaltyRefType             string
	CustomerCode               string
	CompanyCode                string
	Transaction                models.TransactionItem
	AccountPenaltiesDaoService dao.AccountPenaltiesDaoService
	RequestId                  string
}
