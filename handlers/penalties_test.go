package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	. "github.com/smartystreets/goconvey/convey"
)

func buildGetPenaltiesRequest(customerCode string) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, config.CustomerCode, customerCode)
	req := httptest.NewRequest("GET", "/penalties", nil)

	return req.WithContext(ctx)
}

func TestUnitHandleGetPenalties(t *testing.T) {
	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionsMap := &models.AllowedTransactionMap{}

	Convey("Given a request to get penalties", t, func() {
		mockedAccountPenalties := func(params types.AccountPenaltiesParams) (*models.TransactionListResponse, services.ResponseType, error) {
			customerCode := params.CustomerCode
			if customerCode == "INVALID_COMPANY" {
				return nil, services.Error, errors.New("error getting penalties")
			}
			if customerCode == "INVALID_DATA" {
				return nil, services.InvalidData, errors.New("error getting penalties")
			}
			if customerCode == "INTERNAL_SERVER_ERROR" {
				return nil, services.NotFound, errors.New("error getting penalties")
			}
			return nil, services.Success, nil
		}

		mockedGetCompanyCode := func(penaltyRefType string) (string, error) {
			return utils.LateFilingPenaltyCompanyCode, nil
		}

		testCases := []struct {
			companyCode string
			response    int
		}{
			{companyCode: "NI123546", response: http.StatusOK},
			{companyCode: "INVALID_DATA", response: http.StatusBadRequest},
			{companyCode: "INTERNAL_SERVER_ERROR", response: http.StatusInternalServerError},
		}

		getCompanyCode = mockedGetCompanyCode
		accountPenalties = mockedAccountPenalties

		for _, tc := range testCases {

			req := buildGetPenaltiesRequest(tc.companyCode)
			rr := httptest.NewRecorder()

			handler := HandleGetPenalties(nil, penaltyDetailsMap, allowedTransactionsMap)
			handler.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, tc.response)
		}
	})
	Convey("Given a request to get penalties when company code cannot be determined", t, func() {
		getCompanyCode = func(penaltyRefType string) (string, error) {
			return "", errors.New("cannot determine company code")
		}

		rr := httptest.NewRecorder()
		req := buildGetPenaltiesRequest("NI123546")

		handler := HandleGetPenalties(nil, penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req)

		So(rr.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func TestUnitHandleGetPenaltyRefType(t *testing.T) {
	Convey("Get penalty reference type", t, func() {
		testCases := []struct {
			name                   string
			input                  string
			expectedPenaltyRefType string
		}{
			{
				name:                   "Empty",
				input:                  "",
				expectedPenaltyRefType: utils.LateFilingPenaltyRefType,
			},
			{
				name:                   "Late Filing",
				input:                  utils.LateFilingPenaltyRefType,
				expectedPenaltyRefType: utils.LateFilingPenaltyRefType,
			},
			{
				name:                   "Sanctions",
				input:                  utils.SanctionsPenaltyRefType,
				expectedPenaltyRefType: utils.SanctionsPenaltyRefType,
			},
			{
				name:                   "Sanctions ROE",
				input:                  utils.SanctionsRoePenaltyRefType,
				expectedPenaltyRefType: utils.SanctionsRoePenaltyRefType,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				penaltyRefType := GetPenaltyRefType(tc.input)
				Convey(tc.expectedPenaltyRefType, func() {
					So(penaltyRefType, ShouldEqual, tc.expectedPenaltyRefType)
				})
			})
		}
	})
}
