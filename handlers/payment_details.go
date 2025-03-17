package handlers

import (
	"fmt"
	utils2 "github.com/companieshouse/penalty-payment-api/common/utils"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
)

var getCompanyCodeFromTransaction = func(transactions []models.TransactionItem) (string, error) {
	return utils2.GetCompanyCodeFromTransaction(transactions)
}

// HandleGetPaymentDetails retrieves costs for a supplied company number and reference.
func HandleGetPaymentDetails(penaltyDetailsMap *config.PenaltyDetailsMap) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// get payable resource from context, put there by PayableResourceAuthenticationInterceptor
		payableResource, ok := req.Context().Value(config.PayableResource).(*models.PayableResource)
		if !ok {
			log.ErrorR(req, fmt.Errorf("invalid PayableResource in request context"))
			m := models.NewMessageResponse("the payable resource is not present in the request context")
			utils2.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		companyCode, err := getCompanyCodeFromTransaction(payableResource.Transactions)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils2.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		penaltyDetails := penaltyDetailsMap.Details[companyCode]

		// Get the payment details from the payable resource
		paymentDetails, responseType, err := paymentDetailsService.GetPaymentDetailsFromPayableResource(req,
			payableResource, penaltyDetails)
		logData := log.Data{"company_number": payableResource.CompanyNumber, "payable_reference": payableResource.Reference}
		if err != nil {
			switch responseType {
			case services.InvalidData:
				log.DebugR(req, fmt.Sprintf("invalid data getting payment details from payable resource so returning not found [%s]", err.Error()), logData)
				m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
				utils2.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
				return
			default:
				log.ErrorR(req, fmt.Errorf("error when getting payment details from PayableResource: [%v]", err), logData)
				m := models.NewMessageResponse("payable resource does not exist or has insufficient data")
				utils2.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
				return
			}
		}
		utils2.WriteJSON(w, req, paymentDetails)

		log.InfoR(req, "Successful GET request for payment details", logData)
	}
}
