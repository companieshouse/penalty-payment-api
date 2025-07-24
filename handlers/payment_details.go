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
		log.InfoR(req, "start GET payment details request")

		// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
		log.Debug("getting payable resource from context")
		payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)
		if !ok {
			log.ErrorR(req, fmt.Errorf("invalid PayableResource in request context"))
			m := models.NewMessageResponse("the payable resource is not present in the request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}
		log.Debug("got payable resource", log.Data{"payableResource": payableResource})

		penaltyRefType, err := getPenaltyRefTypeFromTransaction(payableResource.Transactions)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		penaltyDetails := penaltyDetailsMap.Details[penaltyRefType]
		log.Debug("penalty details", log.Data{"penaltyDetails": penaltyDetails})

		// Get the payment details from the payable resource
		logData := log.Data{"customer_code": payableResource.CustomerCode, "payable_ref": payableResource.PayableRef}
		log.Info("getting payment details", logData)
		paymentDetails, err := paymentDetailsService.GetPaymentDetailsFromPayableResource(req,
			payableResource, penaltyDetails)
		// can only return either an InvalidData or Success response type
		if err != nil {
			log.DebugR(req, fmt.Sprintf("invalid data getting payment details from payable resource so returning not found [%s]", err.Error()), logData)
			m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}
		log.Debug("got payment details", log.Data{"paymentDetails": paymentDetails})
		utils.WriteJSON(w, req, paymentDetails)

		log.InfoR(req, "GET payment details request completed successfully", logData)
	}
}
