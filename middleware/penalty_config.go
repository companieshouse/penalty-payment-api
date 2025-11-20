package middleware

import (
	"embed"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/configctx"
	"github.com/gorilla/mux"
)

//go:embed assets/*.yaml
var apiFS embed.FS

var (
	penaltyDetails         config.PenaltyDetailsMap
	allowedTransactions    models.AllowedTransactionMap
	penaltyTypesConfig     finance_config.FinancePenaltyTypesConfig
	payablePenaltiesConfig finance_config.FinancePayablePenaltiesConfig
)

func PenaltyConfigMiddleware() mux.MiddlewareFunc {
	const exitErrorFormat = "error configuring service: %s. Exiting"

	log.Info("Loading configuration files")

	b, err := apiFS.ReadFile("assets/penalty_details.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}
	if err := yaml.Unmarshal(b, &penaltyDetails); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	b, err = apiFS.ReadFile("assets/penalty_types.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}
	if err := yaml.Unmarshal(b, &allowedTransactions); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	b, err = finance_config.FS.ReadFile("finance_penalty_types.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	if err := yaml.Unmarshal(b, &penaltyTypesConfig); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	b, err = finance_config.FS.ReadFile("finance_payable_penalties.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	if err := yaml.Unmarshal(b, &payablePenaltiesConfig); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	log.Info("Configuration files loaded successfully")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := configctx.WithConfig(r.Context(), penaltyTypesConfig.FinancePenaltyTypes,
				payablePenaltiesConfig.FinancePayablePenalties, &penaltyDetails, &allowedTransactions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
