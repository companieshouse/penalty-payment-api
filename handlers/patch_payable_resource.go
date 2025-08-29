package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/api"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/service"
)

var (
	handleSendEmailKafkaMessage         = service.SendEmailKafkaMessage
	handlePaymentProcessingKafkaMessage = service.PaymentProcessingKafkaMessage
	wg                                  sync.WaitGroup
	getConfig                           = config.Get
)

// PayResourceHandler will update the resource to mark it as paid and also tell the finance system that the
// transaction(s) associated with it are paid.
func PayResourceHandler(payableResourceService *services.PayableResourceService, e5Client e5.ClientInterface, penaltyPaymentDetails *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := r.Header.Get("X-Request-ID")
		log.InfoC(context, "start PATCH payable resource request")
		// 1. get the payable resource out of the context. authorisation is already handled in the interceptor
		i := r.Context().Value(config.PayableResource)
		if i == nil {
			err := fmt.Errorf("no payable resource in context. check PayableAuthenticationInterceptor is installed")
			log.ErrorC(context, err)
			m := models.NewMessageResponse("no payable request present in request context")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}
		resource := i.(*models.PayableResource)
		logContext := log.Data{"payable_resource": resource}
		log.DebugC(context, "got payable resource from context", logContext)

		// 2. validate the request and check the payment reference against the payment api to validate that it has
		// actually been paid
		log.InfoC(context, "validating request", logContext)
		var request models.PatchResourceRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.ErrorC(context, err)
			m := models.NewMessageResponse("there was a problem reading the request body")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}
		v := validator.New()
		err = v.Struct(request)

		if err != nil {
			log.ErrorC(context, err)
			m := models.NewMessageResponse("the request contained insufficient data and/or failed validation")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}
		log.DebugC(context, "request is valid", log.Data{"request": request})

		log.InfoC(context, "getting payment information", log.Data{"payment_ref": request.Reference, "payable_ref": resource.PayableRef})
		payment, err := service.GetPaymentInformation(request.Reference, r)
		if err != nil {
			log.ErrorC(context, err)
			m := models.NewMessageResponse("the payable resource does not exist")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		log.InfoC(context, "validating payment", log.Data{"payable_ref": resource.PayableRef, "external_payment_id": payment.ExternalPaymentID})
		err = validators.New().ValidateForPayment(*resource, *payment)
		if err != nil {
			log.ErrorC(context, err)
			m := models.NewMessageResponse("there was a problem validating this payment")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}
		log.DebugC(context, "payment is valid", log.Data{"payment": payment})

		wg.Add(3)

		log.InfoC(context, "sending confirmation email", log.Data{"customer_code": resource.CustomerCode, "payable_ref": resource.PayableRef})
		go sendConfirmationEmail(resource, payment, r, w, penaltyPaymentDetails, allowedTransactionsMap, apDaoSvc)
		log.InfoC(context, "updating payable resource as paid", log.Data{"customer_code": resource.CustomerCode, "payable_ref": resource.PayableRef})
		go updateAsPaidInDatabase(resource, payment, payableResourceService, context, w)

		if paymentsProcessingEnabled(context) {
			log.InfoC(context, "payments processing feature enabled")
			go addPaymentsProcessingMsgToTopic(resource, payment, context, w)
		} else {
			log.InfoC(context, "payments processing feature disabled")
			log.InfoC(context, "updating penalty as paid in E5", log.Data{"customer_code": resource.CustomerCode, "payable_ref": resource.PayableRef})
			go updateIssuer(payableResourceService, e5Client, resource, payment, context, w)
		}

		wg.Wait()

		// need to wait to mark the penalty as paid until the go routines above execute as the email
		// sender relies on the state of the penalty in the DB i.e. not paid yet
		log.InfoC(context, "updating account penalty cache record as paid", log.Data{"customer_code": resource.CustomerCode, "payable_ref": resource.PayableRef})
		updateAccountPenaltyAsPaid(resource, apDaoSvc, context)

		log.InfoC(context, "PATCH payable resource request completed successfully", log.Data{"customer_code": resource.CustomerCode})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) // This will not be set if status has already been set
	})
}

func paymentsProcessingEnabled(context string) bool {
	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config for feature flag payments processing enabled, defaulting to false: [%v]", err)
		log.ErrorC(context, err)
		return false
	}
	return cfg.FeatureFlagPaymentsProcessingEnabled
}

func sendConfirmationEmail(resource *models.PayableResource, payment *validators.PaymentInformation, r *http.Request, w http.ResponseWriter,
	penaltyPaymentDetails *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) {

	logContext := log.Data{
		"payable_ref":       resource.PayableRef,
		"payment_reference": payment.Reference,
		"customer_code":     resource.CustomerCode,
		"email_address":     resource.CreatedBy.Email,
	}

	// Send confirmation email
	defer wg.Done()
	err := handleSendEmailKafkaMessage(*resource, r, penaltyPaymentDetails, allowedTransactionsMap, apDaoSvc)
	if err != nil {
		log.ErrorR(r, err, logContext)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.InfoR(r, "Send email kafka message sent", logContext)
}

func updateAsPaidInDatabase(resource *models.PayableResource, payment *validators.PaymentInformation,
	payableResourceService *services.PayableResourceService, context string, w http.ResponseWriter) {
	// Update the payable resource in the db
	defer wg.Done()
	err := payableResourceService.UpdateAsPaid(*resource, *payment)
	if err != nil {
		log.ErrorC(context, err, log.Data{"payable_ref": resource.PayableRef, "payment_reference": payment.Reference})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.InfoC(context, "payment resource is now marked as paid in db", log.Data{
		"payable_ref":   resource.PayableRef,
		"customer_code": resource.CustomerCode,
	})
}

func updateIssuer(payableResourceService *services.PayableResourceService, e5Client e5.ClientInterface, resource *models.PayableResource,
	payment *validators.PaymentInformation, context string, w http.ResponseWriter) {
	// Mark the resource as paid in e5
	defer wg.Done()
	err := api.UpdateIssuerAccountWithPenaltyPaid(payableResourceService, e5Client, *resource, *payment, context)
	if err != nil {
		log.ErrorC(context, err, log.Data{
			"payable_ref":   resource.PayableRef,
			"customer_code": resource.CustomerCode,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.InfoC(context, "successfully initiated process to update payment in E5", log.Data{
		"payable_ref":   resource.PayableRef,
		"customer_code": resource.CustomerCode,
	})
}

func addPaymentsProcessingMsgToTopic(payableResource *models.PayableResource,
	payment *validators.PaymentInformation, context string, w http.ResponseWriter) {
	defer wg.Done()

	logContext := log.Data{
		"created_at":        payableResource.CreatedAt,
		"customer_code":     payableResource.CustomerCode,
		"payable_ref":       payableResource.PayableRef,
		"payment_reference": payment.Reference,
	}
	log.DebugC(context, "adding payments processing message to topic", logContext)
	// send the kafka message to the producer
	err := handlePaymentProcessingKafkaMessage(*payableResource, payment)
	if err != nil {
		log.ErrorC(context, err, logContext)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.InfoC(context, "Payment processing kafka message sent", logContext)
}

func updateAccountPenaltyAsPaid(resource *models.PayableResource, svc dao.AccountPenaltiesDaoService, context string) {
	companyCode, err := getCompanyCodeFromTransaction(resource.Transactions)
	if err != nil {
		log.ErrorC(context, fmt.Errorf("error updating account penalties collection as paid because company code cannot be resolved: [%v]", err),
			log.Data{"customer_code": resource.CustomerCode, "payable_ref": resource.PayableRef})
		return
	}
	penalty := resource.Transactions[0]

	err = svc.UpdateAccountPenaltyAsPaid(resource.CustomerCode, companyCode, penalty.PenaltyRef)
	if err != nil {
		log.ErrorC(context, fmt.Errorf("error updating account penalties collection as paid: [%v]", err),
			log.Data{"customer_code": resource.CustomerCode, "company_code": companyCode,
				"penalty_ref": penalty.PenaltyRef, "payable_ref": resource.PayableRef})
		return
	}

	log.InfoC(context, "account penalties collection has been updated as paid",
		log.Data{"customer_code": resource.CustomerCode, "company_code": companyCode,
			"penalty_ref": penalty.PenaltyRef, "payable_ref": resource.PayableRef})
}
