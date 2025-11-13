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
	"github.com/companieshouse/penalty-payment-api/config"
)

// ConfigResponse represents the JSON response structure
type ConfigResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// HandleConfiguration returns an HTTP handler that serves configuration data
func HandleConfiguration(cfg config.PenaltyConfigProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Create a new slice to hold the extracted structs
		var response []finance_config.PenaltyConfig

		// Loop through the original slice
		for _, p := range cfg.GetPayablePenaltiesConfig() {
			penaltyValue := *p.Penalty
			now := time.Now()

			enabledFrom := penaltyValue.EnabledFrom
			enabledTo := penaltyValue.EnabledTo

			if (enabledFrom != nil && now.After(*enabledFrom)) &&
				(enabledTo == nil || now.Before(*enabledTo)) {
				response = append(response, penaltyValue)
			}
		}

		// Encode response as JSON
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log error and return structured error response
			requestID := r.Header.Get("X-Request-ID")
			log.ErrorC(requestID, fmt.Errorf("error encoding JSON config data: %v", err))
			utils.WriteJSONWithStatus(w, r, models.NewMessageResponse(
				"there was a problem encoding the configuration data"),
				http.StatusInternalServerError)
			return
		}
	}
}
