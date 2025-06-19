package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHandleGetPenalties(t *testing.T) {
	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionsMap := &models.AllowedTransactionMap{}

	Convey("Given a request to get penalties", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
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

		getCompanyCode = mockedGetCompanyCode
		accountPenalties = mockedAccountPenalties

		ctx := context.Background()
		ctx = context.WithValue(ctx, config.CustomerCode, "NI123546")

		req := httptest.NewRequest("GET", "/penalties", nil)
		rr := httptest.NewRecorder()

		handler := HandleGetPenalties(penaltyDetailsMap, allowedTransactionsMap)
		handler.ServeHTTP(rr, req.WithContext(ctx))

		So(rr.Code, ShouldEqual, http.StatusOK)
	})
}
