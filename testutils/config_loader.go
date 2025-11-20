package testutils

import (
	"os"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/configctx"
)

var (
	penaltyConfig configctx.ConfigContext
	once          sync.Once
)

// LoadPenaltyConfigContext loads the config only once and reuses it
func LoadPenaltyConfigContext() configctx.ConfigContext {
	once.Do(func() {
		f, _ := os.ReadFile("../../middleware/assets/penalty_details.yaml")
		var penaltyDetailsMap config.PenaltyDetailsMap
		_ = yaml.Unmarshal(f, &penaltyDetailsMap)

		f, _ = os.ReadFile("../../middleware/assets/penalty_types.yaml")
		var allowedTransactionMap models.AllowedTransactionMap
		_ = yaml.Unmarshal(f, &allowedTransactionMap)

		f, _ = finance_config.FS.ReadFile("finance_penalty_types.yaml")
		var penaltyTypesConfig finance_config.FinancePenaltyTypesConfig
		_ = yaml.Unmarshal(f, &penaltyTypesConfig)

		f, _ = finance_config.FS.ReadFile("finance_payable_penalties.yaml")
		var payablePenaltiesConfig finance_config.FinancePayablePenaltiesConfig
		_ = yaml.Unmarshal(f, &payablePenaltiesConfig)

		penaltyConfig = configctx.ConfigContext{
			PenaltyDetailsMap:     &penaltyDetailsMap,
			AllowedTransactionMap: &allowedTransactionMap,
			PenaltyTypeConfigs:    penaltyTypesConfig.FinancePenaltyTypes,
			PayablePenaltyConfigs: payablePenaltiesConfig.FinancePayablePenalties,
		}
	})
	return penaltyConfig
}
