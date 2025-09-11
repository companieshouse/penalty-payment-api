package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/transformers"
)

var payablePenalty = api.PayablePenalty

// CreatePayableResourceHandler takes a http requests and creates a new payable resource
func CreatePayableResourceHandler(prDaoSvc dao.PayableResourceDaoService, apDaoSvc dao.AccountPenaltiesDaoService,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionMap *models.AllowedTransactionMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get("X-Request-ID")
		log.InfoC(requestId, "start POST payable resource request")

		var request models.PayableRequest
		err := json.NewDecoder(r.Body).Decode(&request)

		userDetails, companyCode, penaltyRefType, failedValidation := extractRequestData(w, r, err, request)
		if failedValidation {
			log.ErrorC(requestId, fmt.Errorf("error extracting request data: %v", err))
			return
		}

		customerCode := r.Context().Value(config.CustomerCode).(string)

		request.CustomerCode = strings.ToUpper(customerCode)
		request.CreatedBy = userDetails
		log.DebugC(requestId, "successfully extracted request data", log.Data{"request": request})

		// Ensure that the transactions in the request are valid payable penalties that exist in E5
		var payablePenalties []models.TransactionItem
		for _, transaction := range request.Transactions {
			params := types.PayablePenaltyParams{
				PenaltyRefType:             penaltyRefType,
				CustomerCode:               customerCode,
				CompanyCode:                companyCode,
				PenaltyDetailsMap:          penaltyDetailsMap,
				Transaction:                transaction,
				AllowedTransactionsMap:     allowedTransactionMap,
				AccountPenaltiesDaoService: apDaoSvc,
				RequestId:                  requestId,
			}
			payablePenalty, err := payablePenalty(params)
			if err != nil {
				log.ErrorC(requestId, fmt.Errorf("invalid request - failed matching against e5"))
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
			log.ErrorC(requestId, fmt.Errorf("invalid request - failed validation"))
			m := models.NewMessageResponse("invalid request body")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}
		log.DebugC(requestId, "request transactions validated, creating payable resource", log.Data{"request": request})

		model := transformers.PayableResourceRequestToDB(&request, requestId)

		err = prDaoSvc.CreatePayableResource(model, requestId)
		if err != nil {
			log.ErrorC(requestId, fmt.Errorf("failed to create payable request in database"))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, r, m, http.StatusInternalServerError)
			return
		}

		payableResource := transformers.PayableResourceDaoToCreatedResponse(model)
		log.DebugC(requestId, "successfully created payable resource", log.Data{"payable_resource": payableResource})

		utils.WriteJSONWithStatus(w, r, payableResource, http.StatusCreated)

		log.InfoC(requestId, "POST payable resource request completed successfully")
	})
}

func extractRequestData(w http.ResponseWriter, r *http.Request, err error, request models.PayableRequest) (authentication.AuthUserDetails, string, string, bool) {
	var authUserDetails authentication.AuthUserDetails
	// request body failed to get decoded
	if err != nil {
		log.ErrorR(r, fmt.Errorf("invalid request"))
		m := models.NewMessageResponse("failed to read request body")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	userDetailsValue := r.Context().Value(authentication.ContextKeyUserDetails)
	if userDetailsValue == nil {
		log.ErrorR(r, fmt.Errorf("user details not in context"))
		m := models.NewMessageResponse("user details not in request context")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	companyCode, err := getCompanyCodeFromTransaction(request.Transactions)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("company code cannot be resolved"))
		m := models.NewMessageResponse("company code cannot be resolved")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	penaltyRefType, err := getPenaltyRefTypeFromTransaction(request.Transactions)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("penalty reference type cannot be resolved"))
		m := models.NewMessageResponse("penalty reference type cannot be resolved")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	authUserDetails = userDetailsValue.(authentication.AuthUserDetails)
	return authUserDetails, companyCode, penaltyRefType, false
}
