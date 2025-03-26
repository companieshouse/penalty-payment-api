package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
)

// HandleGetPayableResource retrieves the payable resource from request context
func HandleGetPayableResource(w http.ResponseWriter, req *http.Request) {
	// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
	payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)

	if !ok {
		log.ErrorR(req, fmt.Errorf("invalid PayableResource in request context"))
		m := models.NewMessageResponse("the payable resource is not present in the request context")
		utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, req, payableResource)
}
