package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/configctx"
)

// HandleGetPaymentDetails retrieves costs for a supplied company number and reference.
func HandleGetPaymentDetails(w http.ResponseWriter, req *http.Request) {
	requestId := log.Context(req)
	log.InfoC(requestId, "start GET payment details request")

	// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
	log.DebugC(requestId, "getting payable resource from context")
	payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)
	if !ok {
		log.ErrorC(requestId, fmt.Errorf("invalid PayableResource in request context"))
		m := models.NewMessageResponse("the payable resource is not present in the request context")
		utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return
	}
	log.DebugC(requestId, "got payable resource", log.Data{"payableResource": payableResource})

	penaltyRefType, err := getPenaltyRefTypeFromTransaction(payableResource.Transactions)
	if err != nil {
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse(err.Error())
		utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return
	}

	penaltyConfig := configctx.FromContext(req.Context())

	penaltyDetails := penaltyConfig.PenaltyDetailsMap.Details[penaltyRefType]
	log.DebugC(requestId, "penalty details", log.Data{"penaltyDetails": penaltyDetails})

	// Get the payment details from the payable resource
	logContext := log.Data{"customer_code": payableResource.CustomerCode, "payable_ref": payableResource.PayableRef}
	log.InfoC(requestId, "getting payment details", logContext)
	paymentDetails, err := paymentDetailsService.GetPaymentDetailsFromPayableResource(req,
		payableResource, penaltyDetails)
	// can only return either an InvalidData or Success response type
	if err != nil {
		log.DebugC(requestId, fmt.Sprintf("invalid data getting payment details from payable resource so returning not found [%s]", err.Error()), logContext)
		m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
		utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
		return
	}
	log.DebugC(requestId, "got payment details", log.Data{"paymentDetails": paymentDetails})
	utils.WriteJSON(w, req, paymentDetails)

	log.InfoC(requestId, "GET payment details request completed successfully")
}
