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
		log.InfoR(req, "start GET penalties request")

		customerCode := req.Context().Value(config.CustomerCode).(string)

		// Determine the CompanyCode from the penaltyRefType which should be on the path
		vars := mux.Vars(req)
		// the penalty reference type is needed further on in the generate_transaction_list to get
		// the ResourceKind from the penalty_details.yaml
		penaltyRefType := GetPenaltyRefType(vars["penalty_reference_type"])
		companyCode, err := getCompanyCode(penaltyRefType)

		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("invalid penalty reference type supplied")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Call service layer to handle request to E5
		transactionListResponse, responseType, err := accountPenalties(penaltyRefType,
			customerCode, companyCode, penaltyDetailsMap, allowedTransactionsMap, apDaoSvc)

		if err != nil {
			log.ErrorR(req, fmt.Errorf("error calling e5 to get transactions: %v", err))
			switch responseType {
			case services.InvalidData:
				m := models.NewMessageResponse("failed to read finance transactions")
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
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
		log.InfoR(req, "GET penalties request completed successfully", log.Data{"customer_code": customerCode})
	}
}

// GetPenaltyRefType gets the penalty reference type from the url vars
// If no penalty reference type is supplied then the request is coming in on the old url
// so defaulting to LateFiling until agreement is made to update other services calling the api
func GetPenaltyRefType(penaltyRefType string) string {
	if len(penaltyRefType) == 0 {
		return utils.LateFilingPenRef
	} else {
		return penaltyRefType
	}
}
