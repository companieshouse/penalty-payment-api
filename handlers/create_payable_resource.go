package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

		request, err := decodeRequest(r)
		if err != nil {
			log.ErrorC(requestId, fmt.Errorf("invalid request"))
			writeJSONResponse(w, r, models.NewMessageResponse("failed to read request body"), http.StatusBadRequest)
			return
		}

		authUserDetails, companyCode, penaltyRefType, failedValidation := extractRequestData(w, r, request)
		if failedValidation {
			log.ErrorC(requestId, fmt.Errorf("error extracting request data"))
			return
		}

		customerCode := r.Context().Value(config.CustomerCode).(string)

		request.CustomerCode = strings.ToUpper(customerCode)
		request.CreatedBy = authUserDetails
		log.DebugC(requestId, "successfully extracted request data", log.Data{"request": request})

		payablePenalties, err := validateTransactions(request.Transactions, penaltyRefType, customerCode, companyCode,
			apDaoSvc, penaltyDetailsMap, allowedTransactionMap, requestId)
		if err != nil {
			log.ErrorC(requestId, fmt.Errorf("invalid request - failed matching against e5"))
			writeJSONResponse(w, r, models.NewMessageResponse("one or more of the transactions you want to pay for do not exist or are not payable at this time"), http.StatusBadRequest)
			return
		}

		request.Transactions = payablePenalties

		err = utils.GetValidator(request)

		if err != nil {
			log.ErrorC(requestId, fmt.Errorf("invalid request - failed validation"))
			writeJSONResponse(w, r, models.NewMessageResponse("invalid request body"), http.StatusBadRequest)
			return
		}

		log.DebugC(requestId, "request transactions validated, creating payable resource", log.Data{"request": request})

		if err := createPayableResource(prDaoSvc, &request, requestId); err != nil {
			log.ErrorC(requestId, fmt.Errorf("failed to create payable request in database"))
			writeJSONResponse(w, r, models.NewMessageResponse("there was a problem handling your request"), http.StatusInternalServerError)
			return
		}

		payableResource := transformers.PayableResourceDaoToCreatedResponse(transformers.PayableResourceRequestToDB(&request, requestId))
		log.DebugC(requestId, "successfully created payable resource", log.Data{"payable_resource": payableResource})

		writeJSONResponse(w, r, payableResource, http.StatusCreated)

		log.InfoC(requestId, "POST payable resource request completed successfully")
	})
}

// decodeRequest decodes the request body into PayableRequest struct
func decodeRequest(r *http.Request) (models.PayableRequest, error) {
	var request models.PayableRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	return request, err
}

// extractRequestData extracts auth details, company code and penalty ref type from the request/context
func extractRequestData(w http.ResponseWriter, r *http.Request, request models.PayableRequest) (authentication.AuthUserDetails, string, string, bool) {
	var authUserDetails authentication.AuthUserDetails

	userDetailsValue := r.Context().Value(authentication.ContextKeyUserDetails)
	if userDetailsValue == nil {
		log.ErrorR(r, fmt.Errorf("user details not in context"))
		writeJSONResponse(w, r, models.NewMessageResponse("user details not in request context"), http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	companyCode, err := getCompanyCodeFromTransaction(request.Transactions)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("company code cannot be resolved"))
		writeJSONResponse(w, r, models.NewMessageResponse("company code cannot be resolved"), http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	penaltyRefType, err := getPenaltyRefTypeFromTransaction(request.Transactions)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("penalty reference type cannot be resolved"))
		writeJSONResponse(w, r, models.NewMessageResponse("penalty reference type cannot be resolved"), http.StatusBadRequest)
		return authUserDetails, "", "", true
	}

	authUserDetails = userDetailsValue.(authentication.AuthUserDetails)
	return authUserDetails, companyCode, penaltyRefType, false
}

// validateTransactions checks that transactions are valid payable penalties
func validateTransactions(transactions []models.TransactionItem, penaltyRefType, customerCode, companyCode string,
	apDaoSvc dao.AccountPenaltiesDaoService,
	penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionMap *models.AllowedTransactionMap,
	requestId string) ([]models.TransactionItem, error) {

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

// createPayableResource saves the payable resource to the database
func createPayableResource(prDaoSvc dao.PayableResourceDaoService, request *models.PayableRequest, requestId string) error {
	model := transformers.PayableResourceRequestToDB(request, requestId)
	return prDaoSvc.CreatePayableResource(model, requestId)
}

// writeJSONResponse is a helper to write JSON response with status code
func writeJSONResponse(w http.ResponseWriter, r *http.Request, payload interface{}, status int) {
	utils.WriteJSONWithStatus(w, r, payload, status)
}
