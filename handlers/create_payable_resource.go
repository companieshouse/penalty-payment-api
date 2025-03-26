package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/transformers"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
)

// CreatePayableResourceHandler takes a http requests and creates a new payable resource
func CreatePayableResourceHandler(svc dao.Service, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionMap *models.AllowedTransactionMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request models.PayableRequest
		err := json.NewDecoder(r.Body).Decode(&request)

		// request body failed to get decoded
		if err != nil {
			log.ErrorR(r, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse("failed to read request body")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		userDetails := r.Context().Value(authentication.ContextKeyUserDetails)
		if userDetails == nil {
			log.ErrorR(r, fmt.Errorf("user details not in context"))
			m := models.NewMessageResponse("user details not in request context")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		companyNumber := r.Context().Value(config.CompanyNumber).(string)

		companyCode, err := getCompanyCodeFromTransaction(request.Transactions)
		if err != nil {
			log.ErrorR(r, fmt.Errorf("company code cannot be determined"))
			m := models.NewMessageResponse("company code cannot be determined")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		request.CompanyNumber = strings.ToUpper(companyNumber)
		request.CreatedBy = userDetails.(authentication.AuthUserDetails)

		// Ensure that the transactions in the request are valid payable penalties that exist in E5
		var payablePenalties []models.TransactionItem
		for _, transaction := range request.Transactions {
			payablePenalty, err := api.PayablePenalty(request.CompanyNumber, companyCode,
				transaction, penaltyDetailsMap, allowedTransactionMap)
			if err != nil {
				log.ErrorR(r, fmt.Errorf("invalid request - failed matching against e5"))
				m := models.NewMessageResponse("one or more of the transactions you want to pay for do not exist or are not payable at this time")
				utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
				return
			}
			payablePenalties = append(payablePenalties, *payablePenalty)
		}

		// Replace request transactions with payable penalties to include updated values in the request
		request.Transactions = payablePenalties

		v := validator.New()
		err = v.Struct(request)

		if err != nil {
			log.ErrorR(r, fmt.Errorf("invalid request - failed validation"))
			m := models.NewMessageResponse("invalid request body")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		model := transformers.PayableResourceRequestToDB(&request)

		err = svc.CreatePayableResource(model)
		if err != nil {
			log.ErrorR(r, fmt.Errorf("failed to create payable request in database"))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, r, m, http.StatusInternalServerError)
			return
		}

		utils.WriteJSONWithStatus(w, r, transformers.PayableResourceDaoToCreatedResponse(model), http.StatusCreated)
	})
}
