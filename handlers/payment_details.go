package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
)

var getCompanyCodeFromTransaction = utils.GetCompanyCodeFromTransaction

// HandleGetPaymentDetails retrieves costs for a supplied company number and reference.
func HandleGetPaymentDetails(penaltyDetailsMap *config.PenaltyDetailsMap) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
		payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)
		if !ok {
			log.ErrorR(req, fmt.Errorf("invalid PayableResource in request context"))
			m := models.NewMessageResponse("the payable resource is not present in the request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		companyCode, err := getCompanyCodeFromTransaction(payableResource.Transactions)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		penaltyDetails := penaltyDetailsMap.Details[companyCode]

		// Get the payment details from the payable resource
		paymentDetails, responseType, err := paymentDetailsService.GetPaymentDetailsFromPayableResource(req,
			payableResource, penaltyDetails)
		logData := log.Data{"customer_code": payableResource.CustomerCode, "payable_ref": payableResource.PayableRef}
		if err != nil {
			switch responseType {
			case services.InvalidData:
				log.DebugR(req, fmt.Sprintf("invalid data getting payment details from payable resource so returning not found [%s]", err.Error()), logData)
				m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
				utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
				return
			default:
				log.ErrorR(req, fmt.Errorf("error when getting payment details from PayableResource: [%v]", err), logData)
				m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
				utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
				return
			}
		}
		utils.WriteJSON(w, req, paymentDetails)

		log.InfoR(req, "Successful GET request for payment details", logData)
	}
}
