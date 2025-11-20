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
		router := mux.NewRouter()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
		Register(router, &config.Config{}, mockPrDaoSvc, mockApDaoSvc)

		healthCheckPath, _ := router.GetRoute("healthcheck").GetPathTemplate()
		healthFinanceCheckPath, _ := router.GetRoute("healthcheck-finance-system").GetPathTemplate()
		getPenaltiesPath, _ := router.GetRoute("get-penalties").GetPathTemplate()
		getPenaltiesOriginalPath, _ := router.GetRoute("get-penalties-legacy").GetPathTemplate()
		createPayablePath, _ := router.GetRoute("create-payable").GetPathTemplate()
		getPayablePath, _ := router.GetRoute("get-payable").GetPathTemplate()
		getPaymentDetailsPath, _ := router.GetRoute("get-payment-details").GetPathTemplate()
		markAsPaidPath, _ := router.GetRoute("mark-as-paid").GetPathTemplate()

		So(healthCheckPath, ShouldEqual, "/penalty-payment-api/healthcheck")
		So(healthFinanceCheckPath, ShouldEqual, "/penalty-payment-api/healthcheck/finance-system")
		So(getPenaltiesPath, ShouldEqual, "/company/{customer_code}/penalties/{penalty_reference_type}")
		So(getPenaltiesOriginalPath, ShouldEqual, "/company/{customer_code}/penalties/late-filing")
		So(createPayablePath, ShouldEqual, "/company/{customer_code}/penalties/payable")
		So(getPayablePath, ShouldEqual, "/company/{customer_code}/penalties/payable/{payable_ref}")
		So(getPaymentDetailsPath, ShouldEqual, "/company/{customer_code}/penalties/payable/{payable_ref}/payment")
		So(markAsPaidPath, ShouldEqual, "/company/{customer_code}/penalties/payable/{payable_ref}/payment")
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
