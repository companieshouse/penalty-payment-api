package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
)

// HandleHealthCheckFinanceSystem checks whether the e5 system is available to take requests
func HandleHealthCheckFinanceSystem(w http.ResponseWriter, r *http.Request) {
	requestId := log.Context(r)

	systemAvailableTime, systemUnavailable, parseError := api.CheckScheduledMaintenance(requestId)

	if parseError {
		log.ErrorC(requestId, fmt.Errorf("parseError from CheckScheduledMaintenance: [%v]", parseError))
		m := models.NewMessageResponse("failed to check scheduled maintenance")
		utils.WriteJSONWithStatus(w, r, m, http.StatusInternalServerError)
		return
	}

	if systemUnavailable {
		m := models.NewMessageTimeResponse("UNHEALTHY - PLANNED MAINTENANCE", systemAvailableTime)
		utils.WriteJSONWithStatus(w, r, m, http.StatusServiceUnavailable)
		log.TraceC(requestId, "Planned maintenance")
		return
	}

	m := models.NewMessageResponse("HEALTHY")
	utils.WriteJSON(w, r, m)
}
