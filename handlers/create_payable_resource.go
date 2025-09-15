package handlers

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/transformers"
)

// CreatePayableResourceHandler takes a http requests and creates a new payable resource
func CreatePayableResourceHandler(prDaoSvc dao.PayableResourceDaoService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the request ID and payable resource request from the context put there by
		// the middleware, PayableResourceRequestValidator
		requestId := r.Context().Value(config.RequestId).(string)
		request := r.Context().Value(config.CreatePayableResource).(models.PayableRequest)

		model := transformers.PayableResourceRequestToDB(&request, requestId)

		err := prDaoSvc.CreatePayableResource(model, requestId)
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
