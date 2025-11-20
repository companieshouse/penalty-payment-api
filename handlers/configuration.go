package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/configctx"
)

type ConfigResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// HandleConfiguration returns an HTTP handler that serves configuration data
func HandleConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response []finance_config.PenaltyConfig

	penaltyConfig := configctx.FromContext(r.Context())

	for _, p := range penaltyConfig.PayablePenaltyConfigs {
		penaltyValue := *p.Penalty
		now := time.Now()

		enabledFrom := penaltyValue.EnabledFrom
		enabledTo := penaltyValue.EnabledTo

		if (enabledFrom != nil && !now.Before(*enabledFrom)) &&
			(enabledTo == nil || now.Before(*enabledTo)) {
			response = append(response, penaltyValue)
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		requestID := r.Header.Get("X-Request-ID")
		log.ErrorC(requestID, fmt.Errorf("error encoding JSON config data: %v", err))
		utils.WriteJSONWithStatus(w, r, models.NewMessageResponse(
			"there was a problem encoding the configuration data"),
			http.StatusInternalServerError)
		return
	}
	return
}
