package private

import (
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitDefaultReasonProvider_GetReason(t *testing.T) {
	Convey("Get reason", t, func() {

		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name   string
			args   args
			reason string
		}{
			{
				name: "Late filing of accounts",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.LateFilingPenaltyCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: "C1",
				}},
				reason: LateFilingPenaltyReason,
			},
			{
				name: "Failure to file a confirmation statement",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsConfirmationStatementTransactionSubType,
				}},
				reason: SanctionsConfirmationStatementReason,
			},
			{
				name: "Failure to deliver a confirmation statement together with the verification statement(s)",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsFailedToVerifyIdentityTransactionSubType,
				}},
				reason: SanctionsFailedToVerifyIdentityReason,
			},
			{
				name: "Sanctions Penalty - Unknown",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: "S2",
				}},
				reason: PenaltyReason,
			},
			{
				name: "Failure to update the Register of Overseas Entities",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsRoeFailureToUpdateTransactionSubType,
				}},
				reason: SanctionsRoeFailureToUpdateReason,
			},
			{
				name: "Other Transaction",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    "5",
					TransactionSubType: "02",
				}},
				reason: "",
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				provider := &DefaultReasonProvider{}
				penaltyTypeConfigs := []finance_config.FinancePenaltyTypeConfig{
					{
						TransactionType:    tc.args.penalty.TransactionType,
						TransactionSubtype: tc.args.penalty.TransactionSubType,
						Reason:             tc.reason,
					},
				}
				got := provider.GetReason(tc.args.penalty, penaltyTypeConfigs)
				So(got, ShouldEqual, tc.reason)
			})
		}
	})
}
