package handlers

import (
	"net/http"

	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/interceptors"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/service"

	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/gorilla/mux"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/middleware"
)

var payableResourceService *services.PayableResourceService
var paymentDetailsService *service.PaymentDetailsService

// Register defines the route mappings for the main router and it's subrouters
func Register(mainRouter *mux.Router, cfg *config.Config, prDaoService dao.PayableResourceDaoService,
	apDaoService dao.AccountPenaltiesDaoService, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) {

	payableResourceService = &services.PayableResourceService{
		Config: cfg,
		DAO:    prDaoService,
	}

	paymentDetailsService = &service.PaymentDetailsService{
		PayableResourceService: payableResourceService,
	}

	payableAuthInterceptor := interceptors.PayableAuthenticationInterceptor{
		Service: *payableResourceService,
	}

	// only oauth2 users can create payable resources
	oauth2OnlyInterceptor := &authentication.OAuth2OnlyAuthenticationInterceptor{
		StrictPaths: map[string][]string{
			"/company/{customer_code}/penalties/payable": {http.MethodPost},
		},
	}

	e5Client := e5.NewClient(cfg.E5Username, cfg.E5APIURL)

	userAuthInterceptor := &authentication.UserAuthenticationInterceptor{
		AllowAPIKeyUser:                true,
		RequireElevatedAPIKeyPrivilege: true,
	}

	mainRouter.HandleFunc("/penalty-payment-api/healthcheck", healthCheck).Methods(http.MethodGet).Name("healthcheck")
	mainRouter.HandleFunc("/penalty-payment-api/healthcheck/finance-system", HandleHealthCheckFinanceSystem).Methods(http.MethodGet).Name("healthcheck-finance-system")

	appRouter := mainRouter.PathPrefix("/company/{customer_code}").Subrouter()
	appRouter.HandleFunc("/penalties/late-filing", HandleGetPenalties(apDaoService, penaltyDetailsMap, allowedTransactionsMap)).Methods(http.MethodGet).Name("get-penalties-legacy")
	appRouter.HandleFunc("/penalties/{penalty_reference_type}", HandleGetPenalties(apDaoService, penaltyDetailsMap, allowedTransactionsMap)).Methods(http.MethodGet).Name("get-penalties")
	appRouter.Handle("/penalties/payable", CreatePayableResourceHandler(prDaoService, apDaoService, penaltyDetailsMap, allowedTransactionsMap)).Methods(http.MethodPost).Name("create-payable")
	appRouter.Use(
		oauth2OnlyInterceptor.OAuth2OnlyAuthenticationIntercept,
		userAuthInterceptor.UserAuthenticationIntercept,
		middleware.CompanyMiddleware,
	)

	// sub router for handling interactions with existing payable resources to apply relevant
	// PayableAuthenticationInterceptor
	existingPayableRouter := appRouter.PathPrefix("/penalties/payable/{payable_ref}").Subrouter()
	existingPayableRouter.HandleFunc("", HandleGetPayableResource).Name("get-payable").Methods(http.MethodGet)
	existingPayableRouter.HandleFunc("/payment", HandleGetPaymentDetails(penaltyDetailsMap)).Methods(http.MethodGet).Name("get-payment-details")
	existingPayableRouter.Use(payableAuthInterceptor.PayableAuthenticationIntercept)

	// separate router for the patch request so that we can apply the interceptor to it without interfering with
	// other routes
	payResourceRouter := appRouter.PathPrefix("/penalties/payable/{payable_ref}/payment").Methods(http.MethodPatch).Subrouter()
	payResourceRouter.Use(payableAuthInterceptor.PayableAuthenticationIntercept, authentication.ElevatedPrivilegesInterceptor)
	payResourceRouter.Handle("", PayResourceHandler(payableResourceService, e5Client, penaltyDetailsMap, allowedTransactionsMap, apDaoService)).Name("mark-as-paid")

	// Set middleware across all routers and sub routers
	mainRouter.Use(log.Handler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
