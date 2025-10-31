package private

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
)

const (
	LateFilingPenaltyReason               = "Late filing of accounts"
	SanctionsConfirmationStatementReason  = "Failure to file a confirmation statement"
	SanctionsFailedToVerifyIdentityReason = "Failed to verify identity"
	SanctionsRoeFailureToUpdateReason     = "Failure to update the Register of Overseas Entities"
	PenaltyReason                         = "Penalty"
)

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
	switch transaction.TransactionSubType {
	case SanctionsConfirmationStatementTransactionSubType:
		return SanctionsConfirmationStatementReason
	case SanctionsFailedToVerifyIdentityTransactionSubType:
		return SanctionsFailedToVerifyIdentityReason
	case SanctionsRoeFailureToUpdateTransactionSubType:
		return SanctionsRoeFailureToUpdateReason
	default:
		return PenaltyReason
	}
}
