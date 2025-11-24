package testutils

import (
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api/configctx"
)

var (
	penaltyConfig configctx.ConfigContext
	once          sync.Once
)

// LoadPenaltyConfigContext loads the config only once and reuses it
func LoadPenaltyConfigContext() configctx.ConfigContext {
	once.Do(func() {
		f, _ := finance_config.FS.ReadFile("finance_penalty_types.yaml")
		var penaltyTypes finance_config.FinancePenaltyTypesConfig
		_ = yaml.Unmarshal(f, &penaltyTypes)

		f, _ = finance_config.FS.ReadFile("finance_payable_penalties.yaml")
		var payablePenalties finance_config.FinancePayablePenaltiesConfig
		_ = yaml.Unmarshal(f, &payablePenalties)

		penaltyConfig = configctx.ConfigContext{
			PenaltyTypes:     penaltyTypes.FinancePenaltyTypes,
			PayablePenalties: payablePenalties.FinancePayablePenalties,
		}
	})
	return penaltyConfig
}
