package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/configctx"
	"github.com/companieshouse/penalty-payment-api/handlers"
	"github.com/companieshouse/penalty-payment-api/testutils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleConfiguration(t *testing.T) {

	Convey("Given a request with ConfigContext containing enabled penalties", t, func() {
		req := httptest.NewRequest(http.MethodGet, "/config", nil)
		res := httptest.NewRecorder()

		penaltyConfig := testutils.LoadPenaltyConfigContext()

		ctxWithConfig := configctx.WithConfig(
			req.Context(),
			penaltyConfig.PenaltyTypeConfigs,
			penaltyConfig.PayablePenaltyConfigs,
			penaltyConfig.PenaltyDetailsMap,
			penaltyConfig.AllowedTransactionMap,
		)
		req = req.WithContext(ctxWithConfig)

		Convey("When HandleConfiguration is called", func() {
			handlers.HandleConfiguration(res, req)

			Convey("Then the response should be 200 OK", func() {
				So(res.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And the response should contain the enabled penalty", func() {
				var response []finance_config.PenaltyConfig
				err := json.NewDecoder(res.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(len(response), ShouldEqual, 2)
				So(response[0].Reason, ShouldEqual, "Late filing of accounts")
			})
		})
	})

	Convey("Given a request with ConfigContext containing no enabled penalties", t, func() {
		now := time.Now()
		enabledFrom := now.Add(2 * time.Hour)
		enabledTo := now.Add(3 * time.Hour)

		mockPenalty := &finance_config.PenaltyConfig{
			Reason:      "Future penalty",
			EnabledFrom: &enabledFrom,
			EnabledTo:   &enabledTo,
		}

		mockPayablePenaltyConfigs := []finance_config.FinancePayablePenaltyConfig{
			{Penalty: mockPenalty},
		}

		req := httptest.NewRequest(http.MethodGet, "/config", nil)
		res := httptest.NewRecorder()

		ctxWithConfig := configctx.WithConfig(
			req.Context(),
			[]finance_config.FinancePenaltyTypeConfig{},
			mockPayablePenaltyConfigs,
			&config.PenaltyDetailsMap{},
			&models.AllowedTransactionMap{},
		)
		req = req.WithContext(ctxWithConfig)

		Convey("When HandleConfiguration is called", func() {
			handlers.HandleConfiguration(res, req)

			Convey("Then the response should be 200 OK", func() {
				So(res.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And the response should contain an empty array", func() {
				var response []finance_config.PenaltyConfig
				err := json.NewDecoder(res.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(len(response), ShouldEqual, 0)
			})
		})
	})
}
