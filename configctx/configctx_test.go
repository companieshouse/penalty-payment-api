package configctx_test

import (
	"context"
	"testing"

	"github.com/companieshouse/penalty-payment-api/configctx"
	"github.com/companieshouse/penalty-payment-api/testutils"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigContext(t *testing.T) {
	Convey("Given a base context", t, func() {
		ctx := context.Background()

		Convey("When WithConfig is called with valid configs", func() {
			penaltyConfig := testutils.LoadPenaltyConfigContext()

			ctxWithConfig := configctx.WithConfig(ctx,
				penaltyConfig.PenaltyTypeConfigs,
				penaltyConfig.PayablePenaltyConfigs,
				penaltyConfig.PenaltyDetailsMap,
				penaltyConfig.AllowedTransactionMap)

			Convey("Then FromContext should return a non-nil ConfigContext", func() {
				cfg := configctx.FromContext(ctxWithConfig)
				So(cfg, ShouldNotBeNil)
				So(cfg.PenaltyTypeConfigs, ShouldResemble, penaltyConfig.PenaltyTypeConfigs)
				So(cfg.PayablePenaltyConfigs, ShouldResemble, penaltyConfig.PayablePenaltyConfigs)
				So(cfg.PenaltyDetailsMap, ShouldEqual, penaltyConfig.PenaltyDetailsMap)
				So(cfg.AllowedTransactionMap, ShouldEqual, penaltyConfig.AllowedTransactionMap)
			})
		})

		Convey("When FromContext is called on a context without config", func() {
			cfg := configctx.FromContext(ctx)
			Convey("Then it should return nil", func() {
				So(cfg, ShouldBeNil)
			})
		})
	})
}
