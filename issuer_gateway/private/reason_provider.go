package private

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
)

const (
	LateFilingPenaltyReason = "Late filing of accounts"
	PenaltyReason           = "Penalty"
)

var getPenaltyTypesConfig = config.GetPenaltyTypesConfig

type ReasonProvider interface {
	GetReason(transaction *models.AccountPenaltiesDataDao) string
}

type DefaultReasonProvider struct{}

func (provider *DefaultReasonProvider) GetReason(transaction *models.AccountPenaltiesDataDao) string {
	if transaction.TransactionType == InvoiceTransactionType {
		switch transaction.CompanyCode {
		case utils.LateFilingPenaltyCompanyCode:
			return LateFilingPenaltyReason
		case utils.SanctionsCompanyCode:
			return getSanctionsReason(transaction)
		default:
			return PenaltyReason
		}
	}
	return ""
}

func getSanctionsReason(transaction *models.AccountPenaltiesDataDao) string {
	for _, penaltyTypeConfig := range getPenaltyTypesConfig() {
		if penaltyTypeConfig.TransactionSubtype == transaction.TransactionSubType {
			return penaltyTypeConfig.Reason
		}
	}
	return PenaltyReason
}
