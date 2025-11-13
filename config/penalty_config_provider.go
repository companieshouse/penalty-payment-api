package config

import "github.com/companieshouse/penalty-payment-api-core/finance_config"

type ConfigurationProvider struct {
	penaltyConfig *PenaltyConfig
}

func NewConfigurationProvider(cfg *PenaltyConfig) PenaltyConfigProvider {
	return &ConfigurationProvider{penaltyConfig: cfg}
}

func (r *ConfigurationProvider) GetPenaltyTypesConfig() []finance_config.FinancePenaltyTypeConfig {
	return r.penaltyConfig.PenaltyTypesConfig
}

func (r *ConfigurationProvider) GetPayablePenaltiesConfig() []finance_config.FinancePayablePenaltyConfig {
	return r.penaltyConfig.PayablePenaltiesConfig
}
