package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/service"
)

// PayableResourceRequestValidator contains the services and configs needed to validate PayableResource requests
type PayableResourceRequestValidator struct {
	PenaltyDetailsMap      *config.PenaltyDetailsMap
	AllowedTransactionsMap *models.AllowedTransactionMap
	ApDaoService           dao.AccountPenaltiesDaoService
}

var payablePenalty = api.PayablePenalty
var getCompanyCode = utils.GetCompanyCode
var getPaymentInformation = service.GetPaymentInformation

// PayableResourceValidate will intercept enhance and/or validate requests to create payable resource,
// patch payable resource and get payment details from payable resource
func (processor *PayableResourceRequestValidator) PayableResourceValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const PAYABLE_REQUEST_PATH_MARKER = "payable"
		path := r.URL.Path

		// Only requests to payable endpoints will be processed here, any other request will
		// just pass through
		if strings.Contains(path, PAYABLE_REQUEST_PATH_MARKER) {
			httpMethod := r.Method
			requestId := r.Header.Get("X-Request-ID")

			// Store requestId in context to use later in the handler
			ctx := context.WithValue(r.Context(), config.RequestId, requestId)

			r = r.WithContext(ctx)
			if httpMethod == http.MethodGet {
				log.InfoC(requestId, "start GET payment details request")
				next.ServeHTTP(w, r)
			} else if httpMethod == http.MethodPost {
				log.InfoC(requestId, "start POST payable resource request")
				validateCreatePayableResourceRequest(r, w, next, processor.PenaltyDetailsMap, processor.AllowedTransactionsMap, processor.ApDaoService, requestId)
			} else if httpMethod == http.MethodPatch {
				log.InfoC(requestId, "start PATCH payable resource request")
				validatePatchPayableResourceRequest(r, w, next, requestId)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validateCreatePayableResourceRequest(r *http.Request, w http.ResponseWriter, next http.Handler, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionMap *models.AllowedTransactionMap, apDaoService dao.AccountPenaltiesDaoService, requestId string) {
	var request models.PayableRequest

	userDetails, companyCode, penaltyRefType, err := extractRequestData(w, r, &request)
	if err != nil {
		log.ErrorC(requestId, fmt.Errorf("error extracting request data: %v", err))
		return
	}

	customerCode := r.Context().Value(config.CustomerCode).(string)

	request.CustomerCode = strings.ToUpper(customerCode)
	request.CreatedBy = userDetails
	log.DebugC(requestId, "successfully extracted request data", log.Data{
		"company_code": companyCode, "penalty_reference_type": penaltyRefType, "request": request})
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
			AccountPenaltiesDaoService: apDaoService,
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

	// Store request in context to use later in the handler
	ctx := context.WithValue(r.Context(), config.CreatePayableResource, request)

	next.ServeHTTP(w, r.WithContext(ctx))
}

func extractRequestData(w http.ResponseWriter, r *http.Request, request *models.PayableRequest) (authentication.AuthUserDetails, string, string, error) {
	var authUserDetails authentication.AuthUserDetails

	err := json.NewDecoder(r.Body).Decode(request)
	// request body failed to get decoded
	if err != nil {
		m := models.NewMessageResponse("failed to read request body")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", err
	}

	userDetailsValue := r.Context().Value(authentication.ContextKeyUserDetails)
	if userDetailsValue == nil {
		m := models.NewMessageResponse("user details not in request context")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", errors.New("user details not in request context")
	}

	penaltyRefType, err := utils.GetPenaltyRefTypeFromTransaction(request.Transactions)
	if err != nil {
		m := models.NewMessageResponse("penalty reference type cannot be resolved")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", errors.New("penalty reference type cannot be resolved")
	}

	companyCode, err := getCompanyCode(penaltyRefType)
	if err != nil {
		m := models.NewMessageResponse("company code cannot be resolved")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return authUserDetails, "", "", errors.New("company code cannot be resolved")
	}

	authUserDetails = userDetailsValue.(authentication.AuthUserDetails)
	return authUserDetails, companyCode, penaltyRefType, nil
}

func validatePatchPayableResourceRequest(r *http.Request, w http.ResponseWriter, next http.Handler, requestId string) {
	// 1. get the payable resource out of the context. authorisation is already handled in the interceptor
	i := r.Context().Value(config.PayableResource)
	if i == nil {
		err := fmt.Errorf("no payable resource in context. check PayableAuthenticationInterceptor is installed")
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse("no payable request present in request context")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return
	}
	resource := i.(*models.PayableResource)
	logContext := log.Data{"payable_resource": resource}
	log.DebugC(requestId, "got payable resource from context", logContext)

	// 2. validate the request and check the payment reference against the payment api to validate that it has
	// actually been paid
	log.InfoC(requestId, "validating request", logContext)
	var request models.PatchResourceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse("there was a problem reading the request body")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return
	}
	v := validator.New()
	err = v.Struct(request)

	if err != nil {
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse("the request contained insufficient data and/or failed validation")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return
	}
	log.DebugC(requestId, "request is valid", log.Data{"request": request})

	log.InfoC(requestId, "getting payment information", log.Data{"payment_ref": request.Reference, "payable_ref": resource.PayableRef})
	payment, err := getPaymentInformation(request.Reference, r)
	if err != nil {
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse("the payable resource does not exist")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return
	}

	log.InfoC(requestId, "validating payment", log.Data{"payable_ref": resource.PayableRef, "external_payment_id": payment.ExternalPaymentID})
	err = validators.New().ValidateForPayment(*resource, *payment)
	if err != nil {
		log.ErrorC(requestId, err)
		m := models.NewMessageResponse("there was a problem validating this payment")
		utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
		return
	}
	log.DebugC(requestId, "payment is valid", log.Data{"payment": payment})

	// Store payment information in context to use later in the handler
	ctx := context.WithValue(r.Context(), config.PaymentInformation, payment)

	next.ServeHTTP(w, r.WithContext(ctx))
}
