package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHandleGetPenalties(t *testing.T) {
	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionsMap := &models.AllowedTransactionMap{}

	Convey("Given a request to get penalties", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
			if companyNumber == "INVALID_COMPANY" {
				return nil, services.Error, errors.New("error getting penalties")
			}
			if companyNumber == "INVALID_DATA" {
				return nil, services.InvalidData, errors.New("error getting penalties")
			}
			if companyNumber == "INTERNAL_SERVER_ERROR" {
				return nil, services.NotFound, errors.New("error getting penalties")
			}
			return nil, services.Success, nil
		}

		mockedGetCompanyCode := func(penaltyReferenceType string) (string, error) {
			return utils.LateFilingPenalty, nil
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

		for _, testCases := range testCases {
			tc := testCases

			ctx := context.Background()
			ctx = context.WithValue(ctx, config.CustomerCode, tc.companyCode)

			req := httptest.NewRequest("GET", "/penalties", nil)
			rr := httptest.NewRecorder()

			handler := HandleGetPenalties(nil, penaltyDetailsMap, allowedTransactionsMap)
			handler.ServeHTTP(rr, req.WithContext(ctx))

			So(rr.Code, ShouldEqual, tc.response)
		}
	})
	Convey("Given a request to get penalties when company code cannot be determined", t, func() {
		mockedGetCompanyCode := func(penaltyReferenceType string) (string, error) {
			return "", errors.New("cannot determine company code")
		}

		getCompanyCode = mockedGetCompanyCode

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.CustomerCode, "NI123546")

		req := httptest.NewRequest("GET", "/penalties", nil)
		rr := httptest.NewRecorder()

		handler := HandleGetPenalties(nil, penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req.WithContext(ctx))

		So(rr.Code, ShouldEqual, http.StatusBadRequest)
	})
}
