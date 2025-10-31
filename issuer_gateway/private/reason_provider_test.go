package private

import (
	"testing"

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
			name string
			args args
			want string
		}{
			{
				name: "Late filing of accounts",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.LateFilingPenaltyCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: "C1",
				}},
				want: LateFilingPenaltyReason,
			},
			{
				name: "Failure to file a confirmation statement",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsConfirmationStatementTransactionSubType,
				}},
				want: SanctionsConfirmationStatementReason,
			},
			{
				name: "Failed to verify identity",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsFailedToVerifyIdentityTransactionSubType,
				}},
				want: SanctionsFailedToVerifyIdentityReason,
			},
			{
				name: "Sanctions Penalty - Unknown",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: "S2",
				}},
				want: PenaltyReason,
			},
			{
				name: "Failure to update the Register of Overseas Entities",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    InvoiceTransactionType,
					TransactionSubType: SanctionsRoeFailureToUpdateTransactionSubType,
				}},
				want: SanctionsRoeFailureToUpdateReason,
			},
			{
				name: "Other Transaction",
				args: args{penalty: &models.AccountPenaltiesDataDao{
					CompanyCode:        utils.SanctionsCompanyCode,
					TransactionType:    "5",
					TransactionSubType: "02",
				}},
				want: "",
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				provider := &DefaultReasonProvider{}
				got := provider.GetReason(tc.args.penalty)
				So(got, ShouldEqual, tc.want)
			})
		}
	})
}
