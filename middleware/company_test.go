package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCompanyMiddleware(t *testing.T) {
	Convey(" Get Company number from request path ", t, func() {
		testCases := []struct {
			name               string
			input              map[string]string
			expectedStatusCode int
		}{
			{
				name:               "Success",
				input:              map[string]string{"company_number": "NI123456"},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:               "Error no company number",
				input:              map[string]string{},
				expectedStatusCode: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				})
				wrappedHandler := CompanyMiddleware(testHandler)
				vars := tc.input
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req = mux.SetURLVars(req, vars)
				rr := httptest.NewRecorder()
				wrappedHandler.ServeHTTP(rr, req)

				if rr.Code != tc.expectedStatusCode {
					t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tc.expectedStatusCode)
				}
			})
		}
	})
}
