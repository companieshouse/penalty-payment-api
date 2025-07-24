package handlers

import (
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/service"
)

var (
	paymentDetailsService            *service.PaymentDetailsService
	getCompanyCode                   = utils.GetCompanyCode
	getCompanyCodeFromTransaction    = utils.GetCompanyCodeFromTransaction
	getPenaltyRefTypeFromTransaction = utils.GetPenaltyRefTypeFromTransaction
)
