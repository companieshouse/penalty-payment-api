package configctx

import (
	"context"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
)

type ctxKey struct{}

type ConfigContext struct {
	PenaltyTypeConfigs    []finance_config.FinancePenaltyTypeConfig
	PayablePenaltyConfigs []finance_config.FinancePayablePenaltyConfig
	PenaltyDetailsMap     *config.PenaltyDetailsMap
	AllowedTransactionMap *models.AllowedTransactionMap
}

func WithConfig(ctx context.Context, penaltyTypeConfigs []finance_config.FinancePenaltyTypeConfig,
	payablePenaltyConfigs []finance_config.FinancePayablePenaltyConfig,
	penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) context.Context {
	return context.WithValue(ctx, ctxKey{}, &ConfigContext{
		penaltyTypeConfigs,
		payablePenaltyConfigs,
		penaltyDetailsMap,
		allowedTransactionsMap})
}
func FromContext(ctx context.Context) *ConfigContext {
	cfg, _ := ctx.Value(ctxKey{}).(*ConfigContext)
	return cfg
}
