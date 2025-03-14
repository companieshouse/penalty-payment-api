package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitRegisterRoutes(t *testing.T) {
	Convey("Register routes", t, func() {
		penaltyDetailsMap = &config.PenaltyDetailsMap{}
		router := mux.NewRouter()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)
		Register(router, &config.Config{}, mockService, penaltyDetailsMap, allowedTransactionsMap)

		healthCheckPath, _ := router.GetRoute("healthcheck").GetPathTemplate()
		healthFinanceCheckPath, _ := router.GetRoute("healthcheck-finance-system").GetPathTemplate()
		getPenaltiesPath, _ := router.GetRoute("get-penalties").GetPathTemplate()
		getPenaltiesOriginalPath, _ := router.GetRoute("get-penalties-original").GetPathTemplate()
		createPayablePath, _ := router.GetRoute("create-payable").GetPathTemplate()
		getPayablePath, _ := router.GetRoute("get-payable").GetPathTemplate()
		getPaymentDetailsPath, _ := router.GetRoute("get-payment-details").GetPathTemplate()
		markAsPaidPath, _ := router.GetRoute("mark-as-paid").GetPathTemplate()

		So(healthCheckPath, ShouldEqual, "/penalty-payment-api/healthcheck")
		So(healthFinanceCheckPath, ShouldEqual, "/penalty-payment-api/healthcheck/finance-system")
		So(getPenaltiesPath, ShouldEqual, "/company/{company_number}/financial-penalties/{penalty_reference_type}")
		So(getPenaltiesOriginalPath, ShouldEqual, "/company/{company_number}/penalties/late-filing")
		So(createPayablePath, ShouldEqual, "/company/{company_number}/financial-penalties/payable")
		So(getPayablePath, ShouldEqual, "/company/{company_number}/financial-penalties/payable/{payable_id}")
		So(getPaymentDetailsPath, ShouldEqual, "/company/{company_number}/financial-penalties/payable/{payable_id}/payment")
		So(markAsPaidPath, ShouldEqual, "/company/{company_number}/financial-penalties/payable/{payable_id}/payment")
	})
}

func TestUnitGetHealthCheck(t *testing.T) {
	Convey("Get HealthCheck", t, func() {
		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()
		healthCheck(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
