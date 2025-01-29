package middleware

import (
	"context"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/utils"
	"github.com/gorilla/mux"
)

// CompanyMiddleware will intercept the company number in the path and stick it into the context
func CompanyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		companyNumber, err := utils.GetCompanyNumberFromVars(vars)

		details := CompanyDetails{map[string]string{
			"CompanyNumber": companyNumber,
		}}

		if err != nil {
			log.ErrorR(r, err)
			m := models.NewMessageResponse("company number not supplied")
			utils.WriteJSONWithStatus(w, r, m, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), config.CompanyDetails, details)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

type CompanyDetails struct {
	M map[string]string
}

func (v CompanyDetails) Get(key string) string {
	return v.M[key]
}
