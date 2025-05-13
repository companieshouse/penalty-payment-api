package middleware

import (
	"context"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/gorilla/mux"
)

// CompanyMiddleware will intercept the customer code in the path and stick it into the context
func CompanyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		customerCode, err := utils.GetCustomerCodeFromVars(vars)

		if err != nil {
			log.ErrorR(r, err)
			m := models.NewMessageResponse("customer code not supplied")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), config.CustomerCode, customerCode)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
