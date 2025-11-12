package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
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
		// Extract query parameter
		configType := r.URL.Query().Get("type")

		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		var configData interface{}

		// Determine which config to return
		switch configType {
		case "penaltyTypesConfig":
			configData = cfg.GetPenaltyTypesConfig()
		case "payablePenaltiesConfig":
			configData = cfg.GetPayablePenaltiesConfig()
		default:
			m := models.NewMessageResponse("invalid configuration type supplied")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		// Prepare response
		response := ConfigResponse{
			Type: configType,
			Data: configData,
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
