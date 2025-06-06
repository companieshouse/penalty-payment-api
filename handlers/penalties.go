package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
	"github.com/gorilla/mux"
)

var getCompanyCode = utils.GetCompanyCode
var accountPenalties = api.AccountPenalties

// HandleGetPenalties retrieves the penalty details for the supplied customer code from e5
func HandleGetPenalties(apDaoSvc dao.AccountPenaltiesDaoService, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.InfoR(req, "start GET penalties request from e5")

		customerCode := req.Context().Value(config.CustomerCode).(string)

		// Determine the CompanyCode from the penaltyReferenceType which should be on the path
		vars := mux.Vars(req)
		penaltyReferenceType := vars["penalty_reference_type"]
		companyCode, err := getCompanyCode(penaltyReferenceType)

		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("invalid penalty reference type supplied")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Call service layer to handle request to E5
		transactionListResponse, responseType, err := accountPenalties(customerCode, companyCode,
			penaltyDetailsMap, allowedTransactionsMap, apDaoSvc)

		if err != nil {
			log.ErrorR(req, fmt.Errorf("error calling e5 to get transactions: %v", err))
			switch responseType {
			case services.InvalidData:
				m := models.NewMessageResponse("failed to read finance transactions")
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			case services.Error:
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
		log.InfoR(req, "Successfully GET penalties from e5", log.Data{"customer_code": customerCode})
	}
}
