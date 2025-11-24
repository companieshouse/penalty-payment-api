package middleware

import (
	"fmt"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api/configctx"
	"github.com/gorilla/mux"
)

var (
	penaltyTypes     finance_config.FinancePenaltyTypesConfig
	payablePenalties finance_config.FinancePayablePenaltiesConfig
)

func PenaltyConfigMiddleware() mux.MiddlewareFunc {
	const exitErrorFormat = "error configuring service: %s. Exiting"

	log.Info("Loading configuration files")

	b, err := finance_config.FS.ReadFile("finance_penalty_types.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	if err := yaml.Unmarshal(b, &penaltyTypes); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	b, err = finance_config.FS.ReadFile("finance_payable_penalties.yaml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	if err := yaml.Unmarshal(b, &payablePenalties); err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
	}

	log.Info("Configuration files loaded successfully")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := configctx.WithConfig(r.Context(),
				penaltyTypes.FinancePenaltyTypes,
				payablePenalties.FinancePayablePenalties)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
