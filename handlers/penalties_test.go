package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/service"
	"github.com/companieshouse/penalty-payment-api/utils"

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

		mockedGetCompanyCode := func(penaltyReference string) (string, error) {
			return utils.LateFilingPenalty, nil
		}

		getCompanyCode = mockedGetCompanyCode
		getPenalties = mockGetPenalties

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.CompanyNumber, "NI123546")

		req := httptest.NewRequest("GET", "/penalties", nil)
		rr := httptest.NewRecorder()

		handler := HandleGetPenalties(penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req.WithContext(ctx))

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}
