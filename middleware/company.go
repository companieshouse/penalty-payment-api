package middleware

import (
	"context"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/gorilla/mux"
)

// CompanyMiddleware will intercept the company number in the path and stick it into the context
func CompanyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		companyNumber, err := utils.GetCompanyNumberFromVars(vars)

		if err != nil {
			log.ErrorR(r, err)
			m := models.NewMessageResponse("company number not supplied")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), config.CompanyNumber, companyNumber)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
