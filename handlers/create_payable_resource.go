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

		request, err := decodeRequestBody(r)
		if handleError(w, r, err, "failed to read request body", http.StatusBadRequest) {
			log.ErrorC(requestId, fmt.Errorf("error extracting request data: %v", err))
			return
		}

		customerCode, authUserDetails, companyCode, penaltyRefType, err := extractRequestContext(r, request)
		if handleError(w, r, err, err.Error(), http.StatusBadRequest) {
			log.ErrorC(requestId, fmt.Errorf("error extracting request data: %v", err))
			return
		}

		request.CustomerCode = strings.ToUpper(customerCode)
		request.CreatedBy = authUserDetails
		log.DebugC(requestId, "successfully extracted request data", log.Data{"request": request})

		validatedTransactions, err := validatePayableTransactions(request.Transactions, apDaoSvc, penaltyDetailsMap, allowedTransactionMap, companyCode, customerCode, penaltyRefType, requestId)
		if handleError(w, r, err, "one or more of the transactions you want to pay for do not exist or are not payable at this time", http.StatusBadRequest) {
			log.ErrorC(requestId, fmt.Errorf("invalid request - failed matching against e5"))
			return
		}

		request.Transactions = validatedTransactions

		err = validateStruct(request)
		if handleError(w, r, err, "invalid request body", http.StatusBadRequest) {
			log.ErrorC(requestId, fmt.Errorf("invalid request - failed validation"))
			return
		}
		log.DebugC(requestId, "request transactions validated, creating payable resource", log.Data{"request": request})

		model := transformers.PayableResourceRequestToDB(&request, requestId)

		err = prDaoSvc.CreatePayableResource(model, requestId)
		if handleError(w, r, err, "there was a problem handling your request", http.StatusInternalServerError) {
			log.ErrorC(requestId, fmt.Errorf("failed to create payable request in database"))
			return
		}

		payableResource := transformers.PayableResourceDaoToCreatedResponse(model)
		log.DebugC(requestId, "successfully created payable resource", log.Data{"payable_resource": payableResource})

		utils.WriteJSONWithStatus(w, r, payableResource, http.StatusCreated)

		log.InfoC(requestId, "POST payable resource request completed successfully")
	})
}

func decodeRequestBody(r *http.Request) (models.PayableRequest, error) {
	var request models.PayableRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

func extractRequestContext(r *http.Request, request models.PayableRequest) (string, authentication.AuthUserDetails, string, string, error) {
	var authUserDetails authentication.AuthUserDetails

	userDetailsValue := r.Context().Value(authentication.ContextKeyUserDetails)
	if userDetailsValue == nil {
		return "", authUserDetails, "", "", fmt.Errorf("user details not in request context")
	}

	companyCode, err := getCompanyCodeFromTransaction(request.Transactions)
	if err != nil {
		return "", authUserDetails, "", "", fmt.Errorf("company code cannot be resolved")
	}

	penaltyRefType, err := getPenaltyRefTypeFromTransaction(request.Transactions)
	if err != nil {
		return "", authUserDetails, "", "", fmt.Errorf("penalty reference type cannot be resolved")
	}

	customerCode := r.Context().Value(config.CustomerCode).(string)
	authUserDetails = userDetailsValue.(authentication.AuthUserDetails)

	return customerCode, authUserDetails, companyCode, penaltyRefType, nil
}

func validatePayableTransactions(transactions []models.TransactionItem, apDaoSvc dao.AccountPenaltiesDaoService,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionMap *models.AllowedTransactionMap,
	companyCode, customerCode, penaltyRefType, requestId string) ([]models.TransactionItem, error) {

	var payablePenalties []models.TransactionItem
	for _, transaction := range transactions {
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
			return nil, err
		}
		payablePenalties = append(payablePenalties, *payablePenalty)
	}
	return payablePenalties, nil
}

func validateStruct(req interface{}) error {
	v := validator.New()
	return v.Struct(req)
}

func handleError(w http.ResponseWriter, r *http.Request, err error, msg string, status int) bool {
	if err != nil {
		log.ErrorR(r, fmt.Errorf("%s", msg))
		m := models.NewMessageResponse(msg)
		utils.WriteJSONWithStatus(w, r, m, status)
		return true
	}
	return false
}
