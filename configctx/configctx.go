package configctx

import (
	"context"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
)

type ctxKey struct{}

type ConfigContext struct {
	PenaltyTypes     map[string]map[string]finance_config.FinancePenaltyTypeConfig
	PayablePenalties map[string]finance_config.FinancePayablePenaltyConfig
}

func WithConfig(ctx context.Context, penaltyTypes map[string]map[string]finance_config.FinancePenaltyTypeConfig,
	payablePenalties map[string]finance_config.FinancePayablePenaltyConfig) context.Context {
	return context.WithValue(ctx, ctxKey{}, &ConfigContext{
		penaltyTypes,
		payablePenalties})
}
func FromContext(ctx context.Context) *ConfigContext {
	cfg, _ := ctx.Value(ctxKey{}).(*ConfigContext)
	return cfg
}
