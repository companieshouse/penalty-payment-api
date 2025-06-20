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
		mockedAccountPenalties := func(penaltyRefType, companyNumber, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
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
		mockedGetCompanyCode := func(penaltyRefType string) (string, error) {
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
				expectedPenaltyRefType: utils.LateFilingPenRef,
			},
			{
				name:                   "Late Filing",
				input:                  utils.LateFilingPenRef,
				expectedPenaltyRefType: utils.LateFilingPenRef,
			},
			{
				name:                   "Sanctions",
				input:                  utils.SanctionsPenRef,
				expectedPenaltyRefType: utils.SanctionsPenRef,
			},
			{
				name:                   "Sanctions ROE",
				input:                  utils.SanctionsRoePenRef,
				expectedPenaltyRefType: utils.SanctionsRoePenRef,
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
