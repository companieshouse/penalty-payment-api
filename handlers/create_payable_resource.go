package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/companieshouse/penalty-payment-api/middleware"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/dao"
	"github.com/companieshouse/penalty-payment-api/transformers"
	"github.com/companieshouse/penalty-payment-api/utils"
	"github.com/companieshouse/penalty-payment-api/validators"
	"gopkg.in/go-playground/validator.v9"
)

// CreatePayableResourceHandler takes a http requests and creates a new payable resource
func CreatePayableResourceHandler(svc dao.Service, penaltyDetailsMap *config.PenaltyDetailsMap) http.Handler {
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
			utils.WriteJSONWithStatus(w, r, m, http.StatusInternalServerError)
			return
		}

		companyNumber := r.Context().Value(config.CompanyDetails).(middleware.CompanyDetails).Get("CompanyNumber")
		companyCode := r.Context().Value(config.CompanyDetails).(middleware.CompanyDetails).Get("CompanyCode")

		if companyNumber == "" || companyCode == "" {
			log.ErrorR(r, fmt.Errorf("company not in context"))
			m := models.NewMessageResponse("company number not in request context")
			utils.WriteJSONWithStatus(w, r, m, http.StatusInternalServerError)
			return
		}

		request.CompanyNumber = strings.ToUpper(companyNumber)
		request.CreatedBy = userDetails.(authentication.AuthUserDetails)

		// validate that the transactions being requested do exist in E5
		validTransactions, err := validators.TransactionsArePayable(request.CompanyNumber, companyCode, request.Transactions, penaltyDetailsMap)
		if err != nil {
			log.ErrorR(r, fmt.Errorf("invalid request - failed matching against e5"))
			m := models.NewMessageResponse("the transactions you want to pay for do not exist or are not payable at this time")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		// validTransactions contains extra values that have been added from E5 validation so override request body transactions with validated transactions
		request.Transactions = validTransactions

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
