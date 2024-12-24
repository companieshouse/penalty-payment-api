package handlers

import (
	"errors"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/service"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHandleGetPenalties(t *testing.T) {
	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionsMap := &models.AllowedTransactionMap{}

	Convey("Given a request to get penalties", t, func() {
		mockGetPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, service.ResponseType, error) {
			if companyNumber == "INVALID_COMPANY" {
				return nil, service.Error, errors.New("error getting penalties")
			}
			if companyNumber == "INVALID_DATA" {
				return nil, service.InvalidData, errors.New("error getting penalties")
			}
			if companyNumber == "INTERNAL_SERVER_ERROR" {
				return nil, service.NotFound, errors.New("error getting penalties")
			}
			return nil, service.Success, nil
		}

		getPenalties = mockGetPenalties

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
				name:               "Error",
				input:              map[string]string{"company_number": "INVALID_COMPANY"},
				expectedStatusCode: http.StatusOK,
			},
			{
				name:               "Error",
				input:              map[string]string{"company_number": "INVALID_DATA"},
				expectedStatusCode: http.StatusBadRequest,
			},
			{
				name:               "Error",
				input:              map[string]string{"company_number": "INTERNAL_SERVER_ERROR"},
				expectedStatusCode: http.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/penalties", nil)
				vars := tc.input
				req = mux.SetURLVars(req, vars)
				rr := httptest.NewRecorder()

				handler := HandleGetPenalties(penaltyDetailsMap, allowedTransactionsMap)
				handler.ServeHTTP(rr, req)

				if rr.Code != tc.expectedStatusCode {
					t.Errorf("handler returned wrong status code: got %v want 200", rr.Code)
				}

			})
		}
	})

	Convey("Given no company number provided", t, func() {
		mockedGetCompanyNumberFromVars := func(vars map[string]string) (string, error) {
			return "", errors.New("error getting penalties")
		}

		getCompanyNumberFromVars = mockedGetCompanyNumberFromVars

		req := httptest.NewRequest("GET", "/penalties", nil)
		vars := map[string]string{}
		req = mux.SetURLVars(req, vars)
		rr := httptest.NewRecorder()

		handler := HandleGetPenalties(penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want 200", rr.Code)
		}

	})

	Convey("Given no company code provided", t, func() {
		mockedGetCompanyCodeFromVars := func() (string, error) {
			return "", errors.New("error getting penalties")
		}

		getCompanyCodeFromVars = mockedGetCompanyCodeFromVars

		req := httptest.NewRequest("GET", "/penalties", nil)
		vars := map[string]string{
			"company_number": "NI123456",
		}
		req = mux.SetURLVars(req, vars)
		rr := httptest.NewRecorder()

		handler := HandleGetPenalties(penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want 200", rr.Code)
		}

	})
}
