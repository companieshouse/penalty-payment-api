package private

import (
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitDefaultPayableStatusProvider_GetPayableStatus(t *testing.T) {

	Convey("Get open payable status for late filing penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty (valid)",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount and not paid",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN1",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN2",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is PEN3",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN1",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, DCADunningStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN2",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is PEN3",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN1",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN2",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is PEN3",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN1",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN2",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is PEN3",
				args: args{
					penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
						addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for late filing penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Late filing penalty with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount less than 0 and is paid",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(true, -150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is DCA, dunning status is DCA (no trailing spaces)",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, DCAAccountStatus, DCADunningStatus)},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is DCA",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is DCA",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is WDR, dunning status is DCA",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is IPEN1",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is HLD, dunning status is IPEN2",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Late filing penalty with outstanding amount, not paid, account status is CHS, dunning status is CAN",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(false, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("CAN"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions confirmation statement", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions confirmation statement (valid)",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount and not paid",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions confirmation statement", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions confirmation statement with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount less than 0 and is paid",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions confirmation statement with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions failed to verify identity", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions failed to verify identity (valid)",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount and not paid",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions failed to verify identity", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions failed to verify identity with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount less than 0 and is paid",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(true, -75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 75, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get open payable status for sanctions ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: OpenPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: OpenPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed payable status for sanctions ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE with outstanding amount is 0 and is paid (valid)",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount less than 0 and is paid",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(true, -250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status and dunning status is dca",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account not dca and dunning status is dca",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(DCADunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen3",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen3",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen3",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is wdr, dunning status is pen3",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, WDRAccountStatus,
					addTrailingSpacesToDunningStatus(PEN3DunningStatus))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is ipen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN1"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is ipen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN2"))},
				want: ClosedPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is ipen3",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus("IPEN3"))},
				want: ClosedPayableStatus,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &yesterday
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get closed pending allocation payable status for penalty", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
		}{
			{
				name: "Late filing penalty paid today",
				args: args{penalty: buildLateFilingPenaltyTestAccountPenaltiesDataDao(true, 150, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions penalty paid today",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
			{
				name: "Sanctions ROE penalty paid today",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				closedAt := &now
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, closedAt, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, ClosedPendingAllocationPayableStatus)
			})
		}
	})

	Convey("Get closed instalment plan payable status for Late filing penalty with associated instalment plan", t, func() {
		// Given
		lateFilingPaidPenalty := buildPaidPenaltyTransaction("A3784631", "2025-05-02", "2024-12-31", 3000, "2025-05-02")
		e5Transactions := []models.AccountPenaltiesDataDao{
			buildInstalmentTransaction("A3784631-001", "2025-07-23", "2024-12-31", 300, "2024-08-26"),
			buildInstalmentTransaction("A3784631-002", "2025-07-23", "2024-12-31", 300, "2024-09-26"),
			buildInstalmentTransaction("A3784631-003", "2025-07-23", "2024-12-31", 300, "2024-10-26"),
			buildInstalmentTransaction("A3784631-004", "2025-07-23", "2024-12-31", 300, "2024-11-26"),
			buildInstalmentTransaction("A3784631-005", "2025-07-23", "2024-12-31", 300, "2024-12-26"),
			lateFilingPaidPenalty,
			buildInstalmentPlanTransaction("A3784631", "2025-07-23", "2024-12-31", 3000, "2025-07-23"),
			buildInstalmentTransaction("A3784631-006", "2025-07-23", "2024-12-31", 125, "2025-01-26"),
			buildPlanAdjustmentTransaction("A3784631", "2025-10-29", "2024-12-31", -1625, "2025-10-29"),
		}

		// When
		provider := &DefaultPayableStatusProvider{}
		got := provider.GetPayableStatus(types.Penalty.String(), &lateFilingPaidPenalty, nil, e5Transactions, allowedTransactionMap, &cfg)

		// Then
		So(got, ShouldEqual, ClosedInstalmentPlanPayableStatus)
	})

	Convey("Get closed payable status for Late filing penalty without an associated instalment plan", t, func() {
		// Given
		lateFilingPaidPenalty := buildPaidPenaltyTransaction("A3784631", "2025-05-02", "2024-12-31", 3000, "2025-05-02")
		e5Transactions := []models.AccountPenaltiesDataDao{
			buildInstalmentTransaction("A7138463-001", "2024-07-23", "2023-12-31", 300, "2023-08-26"),
			buildInstalmentTransaction("A7138463-002", "2024-07-23", "2023-12-31", 300, "2023-09-26"),
			buildInstalmentTransaction("A7138463-003", "2024-07-23", "2023-12-31", 300, "2023-10-26"),
			buildInstalmentTransaction("A7138463-004", "2024-07-23", "2023-12-31", 300, "2023-11-26"),
			buildInstalmentTransaction("A7138463-005", "2024-07-23", "2023-12-31", 300, "2023-12-26"),
			lateFilingPaidPenalty,
			buildInstalmentPlanTransaction("A7138463", "2024-07-23", "2023-12-31", 3000, "2024-07-23"),
			buildInstalmentTransaction("A7138463-006", "2024-07-23", "2023-12-31", 125, "2024-01-26"),
			buildPlanAdjustmentTransaction("A7138463", "2024-10-29", "2023-12-31", -1625, "2024-10-29"),
		}

		// When
		provider := &DefaultPayableStatusProvider{}
		got := provider.GetPayableStatus(types.Penalty.String(), &lateFilingPaidPenalty, nil, e5Transactions, allowedTransactionMap, &cfg)

		// Then
		So(got, ShouldEqual, ClosedPayableStatus)
	})

	Convey("Get closed instalment plan payable status for Late filing penalty with associated instalment plan", t, func() {
		// Given
		lateFilingPaidPenalty := buildPaidPenaltyTransaction("A3784631", "2025-05-02", "2024-12-31", 3000, "2025-05-02")
		e5Transactions := []models.AccountPenaltiesDataDao{
			buildExhaustedWriteOffTransaction("A3784631-001", "2025-07-23", "2024-12-31", 300, "2024-08-26"),
		}

		// When
		provider := &DefaultPayableStatusProvider{}
		got := provider.GetPayableStatus(types.Penalty.String(), &lateFilingPaidPenalty, nil, e5Transactions, allowedTransactionMap, &cfg)

		// Then
		So(got, ShouldEqual, ClosedPenStrategyExhaustedPayableStatus)
	})

	Convey("Get closed payable status for an existing paid penalty from E5 in the same account as a newly paid penalty", t, func() {
		oldPaidPenalty := buildSanctionsRoeTestAccountPenaltiesDataDao(true, 0, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))
		newPaidPenalty := buildSanctionsRoeTestAccountPenaltiesDataDao(true, 250, CHSAccountStatus,
			addTrailingSpacesToDunningStatus(PEN1DunningStatus))

		closedAt := time.Now()
		e5Transactions := []models.AccountPenaltiesDataDao{*oldPaidPenalty, *newPaidPenalty}
		provider := &DefaultPayableStatusProvider{}
		So(provider.GetPayableStatus(types.Penalty.String(), oldPaidPenalty, &closedAt, e5Transactions, allowedTransactionMap, &cfg), ShouldEqual, ClosedPayableStatus)
		So(provider.GetPayableStatus(types.Penalty.String(), newPaidPenalty, &closedAt, e5Transactions, allowedTransactionMap, &cfg), ShouldEqual, ClosedPendingAllocationPayableStatus)

	})

	Convey("Get disabled payable status for sanctions - Confirmation Statement", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions (valid)",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount and not paid",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsConfirmationStatementTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get disabled payable status for sanctions - Failed to verify identity", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions failed to verify identity (valid)",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount and not paid",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions failed to verify identity with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsFailedToVerifyIdentityTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get disabled payable status for sanctions - ROE", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount and not paid",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid and account on hold",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is dca, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, DCAAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is chs, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen1",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE with outstanding amount, not paid, account status is hld, dunning status is pen2",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, HLDAccountStatus,
					addTrailingSpacesToDunningStatus(PEN2DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsRoeFailureToUpdateTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})

	Convey("Get disabled payable status for sanctions - Multiple Subtypes", t, func() {
		type args struct {
			penalty *models.AccountPenaltiesDataDao
		}
		testCases := []struct {
			name string
			args args
			want string
		}{
			{
				name: "Sanctions (valid)",
				args: args{penalty: buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
			{
				name: "Sanctions ROE (valid)",
				args: args{penalty: buildSanctionsRoeTestAccountPenaltiesDataDao(false, 250, CHSAccountStatus,
					addTrailingSpacesToDunningStatus(PEN1DunningStatus))},
				want: DisabledPayableStatus,
			},
		}
		cfg.DisabledPenaltyTransactionSubtypes = SanctionsMultipleTransactionSubType
		for _, tc := range testCases {
			Convey(tc.name, func() {
				penalty := tc.args.penalty
				provider := &DefaultPayableStatusProvider{}
				got := provider.GetPayableStatus(types.Penalty.String(), penalty, &now, []models.AccountPenaltiesDataDao{*penalty}, allowedTransactionMap, &cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func buildLateFilingPenaltyTestAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.LateFilingPenaltyCompanyCode,
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: "A1234567",
		Amount:               150,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      "1",
		TransactionSubType:   "EU",
		TypeDescription:      "Penalty Ltd Wel & Eng <=1m    LTDWA",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildSanctionsConfirmationStatementTestAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "E1",
		CustomerCode:         "12345678",
		TransactionReference: "P1234567",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      InvoiceTransactionType,
		TransactionSubType:   SanctionsConfirmationStatementTransactionSubType,
		TypeDescription:      "CS01                                    ",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildExhaustedWriteOffTransaction(transactionReference string, transactionDate string, madeUpDate string, amount float64, dueDate string) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          "LP",
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: transactionReference,
		TransactionDate:      transactionDate,
		MadeUpDate:           madeUpDate,
		Amount:               amount,
		OutstandingAmount:    0,
		IsPaid:               true,
		TransactionType:      "4",
		TransactionSubType:   "82",
		TypeDescription:      "W/PEN STRATEGY EXHAUSTED                ",
		DueDate:              dueDate,
		AccountStatus:        "CHS",
		DunningStatus:        "DCA         ",
	}
}

func buildSanctionsFailedToVerifyIdentityTestAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "E1",
		CustomerCode:         "12345678",
		TransactionReference: "P2234567",
		Amount:               75,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      InvoiceTransactionType,
		TransactionSubType:   SanctionsFailedToVerifyIdentityTransactionSubType,
		TypeDescription:      "CS01 IDV                                ",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildSanctionsRoeTestAccountPenaltiesDataDao(isPaid bool, outstandingAmount float64, accountStatus, dunningStatus string) *models.AccountPenaltiesDataDao {
	dataDao := buildTestAccountPenaltiesDataDao(AccountPenaltiesParams{
		CompanyCode:          utils.SanctionsCompanyCode,
		LedgerCode:           "FU",
		CustomerCode:         "OE123456",
		TransactionReference: "U1234567",
		Amount:               250,
		OutstandingAmount:    outstandingAmount,
		IsPaid:               isPaid,
		TransactionType:      InvoiceTransactionType,
		TransactionSubType:   SanctionsRoeFailureToUpdateTransactionSubType,
		TypeDescription:      "PENU                                    ",
		AccountStatus:        accountStatus,
		DunningStatus:        dunningStatus,
	})

	return &dataDao
}

func buildInstalmentTransaction(transactionReference string, transactionDate string, madeUpDate string, amount float64, dueDate string) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          "LP",
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: transactionReference,
		TransactionDate:      transactionDate,
		MadeUpDate:           madeUpDate,
		Amount:               amount,
		OutstandingAmount:    0,
		IsPaid:               true,
		TransactionType:      "1",
		TransactionSubType:   "I1",
		TypeDescription:      "Instalment                              ",
		DueDate:              dueDate,
		AccountStatus:        "CHS",
		DunningStatus:        "IPEN1       ",
	}
}

func buildPaidPenaltyTransaction(transactionReference string, transactionDate string, madeUpDate string, amount float64, dueDate string) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          "LP",
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: transactionReference,
		TransactionDate:      transactionDate,
		MadeUpDate:           madeUpDate,
		Amount:               amount,
		OutstandingAmount:    0,
		IsPaid:               true,
		TransactionType:      "1",
		TransactionSubType:   "EL",
		TypeDescription:      "Double DBL LTD E&W> 6 MNTHS   DLTWD     ",
		DueDate:              dueDate,
		AccountStatus:        "CHS",
		DunningStatus:        "PEN2        ",
	}
}

func buildInstalmentPlanTransaction(transactionReference string, transactionDate string, madeUpDate string, amount float64, dueDate string) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          "LP",
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: transactionReference,
		TransactionDate:      transactionDate,
		MadeUpDate:           madeUpDate,
		Amount:               amount,
		OutstandingAmount:    0,
		IsPaid:               true,
		TransactionType:      "P",
		TransactionSubType:   "00",
		TypeDescription:      "Instalment Plan                         ",
		DueDate:              dueDate,
		AccountStatus:        "CHS",
		DunningStatus:        "            ",
	}
}

func buildPlanAdjustmentTransaction(transactionReference string, transactionDate string, madeUpDate string, amount float64, dueDate string) models.AccountPenaltiesDataDao {
	return models.AccountPenaltiesDataDao{
		CompanyCode:          "LP",
		LedgerCode:           "EW",
		CustomerCode:         "12345678",
		TransactionReference: transactionReference,
		TransactionDate:      transactionDate,
		MadeUpDate:           madeUpDate,
		Amount:               amount,
		OutstandingAmount:    0,
		IsPaid:               true,
		TransactionType:      "7",
		TransactionSubType:   "08",
		TypeDescription:      "Plan Adjustment                         ",
		DueDate:              dueDate,
		AccountStatus:        "CHS",
		DunningStatus:        "            ",
	}
}
