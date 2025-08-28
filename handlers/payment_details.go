package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
)

// HandleGetPaymentDetails retrieves costs for a supplied company number and reference.
func HandleGetPaymentDetails(penaltyDetailsMap *config.PenaltyDetailsMap) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		context := req.Header.Get("X-Request-ID")
		log.InfoC(context, "start GET payment details request")

		// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
		log.DebugC(context, "getting payable resource from context")
		payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)
		if !ok {
			log.ErrorC(context, fmt.Errorf("invalid PayableResource in request context"))
			m := models.NewMessageResponse("the payable resource is not present in the request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}
		log.DebugC(context, "got payable resource", log.Data{"payableResource": payableResource})

		penaltyRefType, err := getPenaltyRefTypeFromTransaction(payableResource.Transactions)
		if err != nil {
			log.ErrorC(context, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		penaltyDetails := penaltyDetailsMap.Details[penaltyRefType]
		log.DebugC(context, "penalty details", log.Data{"penaltyDetails": penaltyDetails})

		// Get the payment details from the payable resource
		logContext := log.Data{"customer_code": payableResource.CustomerCode, "payable_ref": payableResource.PayableRef}
		log.InfoC(context, "getting payment details", logContext)
		paymentDetails, err := paymentDetailsService.GetPaymentDetailsFromPayableResource(req,
			payableResource, penaltyDetails)
		// can only return either an InvalidData or Success response type
		if err != nil {
			log.DebugC(context, fmt.Sprintf("invalid data getting payment details from payable resource so returning not found [%s]", err.Error()), logContext)
			m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}
		log.DebugC(context, "got payment details", log.Data{"paymentDetails": paymentDetails})
		utils.WriteJSON(w, req, paymentDetails)

		log.InfoC(context, "GET payment details request completed successfully")
	}
}
