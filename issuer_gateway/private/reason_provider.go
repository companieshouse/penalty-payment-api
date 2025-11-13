package private

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
)

const (
	LateFilingPenaltyReason               = "Late filing of accounts"
	SanctionsConfirmationStatementReason  = "Failure to file a confirmation statement"
	SanctionsFailedToVerifyIdentityReason = "Failure to deliver a confirmation statement together with the verification statement(s)"
	SanctionsRoeFailureToUpdateReason     = "Failure to update the Register of Overseas Entities"
	PenaltyReason                         = "Penalty"
)

type ReasonProvider interface {
	GetReason(transaction *models.AccountPenaltiesDataDao, configProvider config.PenaltyConfigProvider) string
}

type DefaultReasonProvider struct{}

func (provider *DefaultReasonProvider) GetReason(transaction *models.AccountPenaltiesDataDao,
	configProvider config.PenaltyConfigProvider) string {
	if transaction.TransactionType == InvoiceTransactionType {
		switch transaction.CompanyCode {
		case utils.LateFilingPenaltyCompanyCode:
			return LateFilingPenaltyReason
		case utils.SanctionsCompanyCode:
			return getSanctionsReason(transaction, configProvider)
		default:
			return PenaltyReason
		}
	}
	return ""
}

func getSanctionsReason(transaction *models.AccountPenaltiesDataDao, configProvider config.PenaltyConfigProvider) string {
	for _, penaltyTypeConfig := range configProvider.GetPenaltyTypesConfig() {
		if penaltyTypeConfig.TransactionSubtype == transaction.TransactionSubType {
			return penaltyTypeConfig.Reason
		}
	}
	return PenaltyReason
}
