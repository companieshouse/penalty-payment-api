package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/companieshouse/penalty-payment-api/configctx"
)

func TestPenaltyConfigMiddleware(t *testing.T) {
	Convey("Given a router with PenaltyConfigMiddleware", t, func() {
		router := mux.NewRouter()
		router.Use(PenaltyConfigMiddleware())

		var capturedCfg *configctx.ConfigContext

		router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			cfg := configctx.FromContext(r.Context())
			capturedCfg = cfg
			if cfg == nil {
				http.Error(w, "Config not found", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		Convey("When a request is sent through the middleware", func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			Convey("Then the response should be 200 OK", func() {
				So(rr.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And the context should contain a non-nil ConfigContext", func() {
				So(capturedCfg, ShouldNotBeNil)
				So(capturedCfg.PenaltyDetailsMap, ShouldNotBeNil)
				So(capturedCfg.AllowedTransactionMap, ShouldNotBeNil)
				So(capturedCfg.PenaltyTypeConfigs, ShouldNotBeNil)
				So(capturedCfg.PayablePenaltyConfigs, ShouldNotBeNil)
			})
		})

	})
}
