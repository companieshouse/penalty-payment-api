package private

import (
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
)

const (
	LateFilingPenaltyReason = "Late filing of accounts"
	PenaltyReason           = "Penalty"
)

type ReasonProvider interface {
	GetReason(transaction *models.AccountPenaltiesDataDao, penaltyTypes []finance_config.FinancePenaltyTypeConfig) string
}

type DefaultReasonProvider struct{}

func (provider *DefaultReasonProvider) GetReason(transaction *models.AccountPenaltiesDataDao, penaltyTypes []finance_config.FinancePenaltyTypeConfig) string {
	if transaction.TransactionType == InvoiceTransactionType {
		switch transaction.CompanyCode {
		case utils.LateFilingPenaltyCompanyCode:
			return LateFilingPenaltyReason
		case utils.SanctionsCompanyCode:
			return getSanctionsReason(transaction, penaltyTypes)
		default:
			return PenaltyReason
		}
	}
	return ""
}

func getSanctionsReason(transaction *models.AccountPenaltiesDataDao, penaltyTypes []finance_config.FinancePenaltyTypeConfig) string {

	for _, penaltyTypeConfig := range penaltyTypes {
		if penaltyTypeConfig.TransactionSubtype == transaction.TransactionSubType {
			return penaltyTypeConfig.Reason
		}
	}
	return PenaltyReason
}
