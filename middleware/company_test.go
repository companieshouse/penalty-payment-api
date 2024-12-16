package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCompanyMiddleware(t *testing.T) {
	Convey("Test Company Middleware", t, func() {

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		wrappedHandler := CompanyMiddleware(testHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recorder, req)

		resp := recorder.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, but got %v", resp.StatusCode)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, "OK") {
			t.Errorf("Expected body 'OK', but got %v", body)
		}
	})

}
