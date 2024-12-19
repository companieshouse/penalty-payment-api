package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/service"
	"github.com/companieshouse/penalty-payment-api/utils"
	"github.com/gorilla/mux"
)

var getCompanyNumberFromVars = func(vars map[string]string) (string, error) {
	return utils.GetCompanyNumberFromVars(vars)
}
var getCompanyCodeFromVars = func() (string, error) {
	return utils.GetCompanyCodeFromVars()
}
var getPenalties = func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, service.ResponseType, error) {
	return service.GetPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionsMap)
}

// HandleGetPenalties retrieves the penalty details for the supplied company number from e5
func HandleGetPenalties(penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.InfoR(req, "start GET penalties request from e5")

		// Check for a company number in request
		vars := mux.Vars(req)
		// only company number in the route variables

		companyNumber, err := getCompanyNumberFromVars(vars)

		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("company number is not in request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		companyCode, err := getCompanyCodeFromVars()
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("company code is not in request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		companyNumber = strings.ToUpper(companyNumber)
		companyCode = strings.ToUpper(companyCode)

		// Call service layer to handle request to E5
		transactionListResponse, responseType, err := getPenalties(companyNumber, companyCode, penaltyDetailsMap, allowedTransactionsMap)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error calling e5 to get transactions: %v", err))
			switch responseType {
			case service.InvalidData:
				m := models.NewMessageResponse("failed to read finance transactions")
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			case service.Error:
			default:
				m := models.NewMessageResponse("there was a problem communicating with the finance backend")
				utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
				return
			}
		}
		// response body contains fully decorated REST model
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(transactionListResponse)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error writing response: %v", err))
			return
		}
		log.InfoR(req, "Successfully GET penalties from e5", log.Data{"company_number": companyNumber})
	}
}
