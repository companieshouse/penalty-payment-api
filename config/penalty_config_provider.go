package config

import "github.com/companieshouse/penalty-payment-api-core/finance_config"

type ConfigurationProvider struct{}

func NewConfigurationProvider() PenaltyConfigProvider {
	return &ConfigurationProvider{}
}

func (r *ConfigurationProvider) GetPenaltyTypesConfig() []finance_config.FinancePenaltyTypeConfig {
	return penaltyConfig.PenaltyTypesConfig
}

func (r *ConfigurationProvider) GetPayablePenaltiesConfig() []finance_config.FinancePayablePenaltyConfig {
	return penaltyConfig.PayablePenaltiesConfig
}
