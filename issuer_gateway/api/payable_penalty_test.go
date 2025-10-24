package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func accountPenaltiesResponse(unpaidPenaltyCount int) *models.TransactionListResponse {

	unpaidPenalty := models.TransactionListItem{
		Etag:            "etag",
		Kind:            "penalty",
		IsPaid:          false,
		IsDCA:           false,
		DueDate:         "2018-05-14",
		MadeUpDate:      "2017-06-30",
		TransactionDate: "2018-04-30",
		OriginalAmount:  150,
		Outstanding:     150,
		Type:            "penalty",
		Reason:          "Late filing of accounts",
		PayableStatus:   "OPEN",
	}

	paidPenalty := unpaidPenalty
	paidPenalty.ID = "00482775"
	paidPenalty.IsPaid = true
	paidPenalty.PayableStatus = "CLOSED"

	unpaidOther := unpaidPenalty
	unpaidOther.ID = "00482776"
	unpaidOther.Type = "other"

	response := models.TransactionListResponse{
		Items: []models.TransactionListItem{paidPenalty, unpaidOther},
	}

	for i := 1; i <= unpaidPenaltyCount; i++ {
		item := unpaidPenalty
		item.ID = fmt.Sprintf("A%07d", i)
		response.Items = append(response.Items, item)
	}

	return &response
}

func generateParams(daoService dao.AccountPenaltiesDaoService, transaction models.TransactionItem) types.PayablePenaltyParams {
	return types.PayablePenaltyParams{
		PenaltyRefType:             utils.LateFilingPenaltyRefType,
		CompanyCode:                utils.LateFilingPenaltyCompanyCode,
		CustomerCode:               "10000024",
		PenaltyDetailsMap:          &config.PenaltyDetailsMap{},
		Transaction:                transaction,
		RequestId:                  "",
		AccountPenaltiesDaoService: daoService,
	}

}

func TestUnitPayablePenalty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)

	Convey("error is returned when fetching account penalties fails", t, func() {
		accountPenaltiesErr := errors.New("failed to fetch account penalties")
		getAccountPenalties = func(params types.AccountPenaltiesParams) (*models.TransactionListResponse, services.ResponseType, error) {
			return nil, services.Error, accountPenaltiesErr
		}

		transaction := models.TransactionItem{PenaltyRef: "121"}
		payablePenalty, err := PayablePenalty(generateParams(mockApDaoSvc, transaction))

		So(payablePenalty, ShouldBeNil)
		So(err, ShouldEqual, accountPenaltiesErr)
	})

	Convey("payable penalty is successfully returned for multiple unpaid penalties", t, func() {
		getAccountPenalties = func(params types.AccountPenaltiesParams) (*models.TransactionListResponse, services.ResponseType, error) {
			return accountPenaltiesResponse(2), services.Success, nil
		}

		unpaidPenalty := models.TransactionItem{
			PenaltyRef: "A0000002",
			Amount:     150,
			Type:       "penalty",
			MadeUpDate: "2017-06-30",
			IsDCA:      false,
			IsPaid:     false,
			Reason:     "Late filing of accounts",
		}
		gotPayablePenalty, err := PayablePenalty(generateParams(mockApDaoSvc, unpaidPenalty))

		So(gotPayablePenalty, ShouldResemble, &unpaidPenalty)
		So(err, ShouldBeNil)
	})

	Convey("payable penalty is successfully returned", t, func() {
		getAccountPenalties = func(params types.AccountPenaltiesParams) (*models.TransactionListResponse, services.ResponseType, error) {
			return accountPenaltiesResponse(1), services.Success, nil
		}

		wantPayablePenalty := &models.TransactionItem{
			PenaltyRef: "121",
			Amount:     150,
			Type:       "penalty",
			MadeUpDate: "2017-06-30",
			IsDCA:      false,
			IsPaid:     false,
		}
		getMatchingPenalty = func(referenceTransactions []models.TransactionListItem, transactionToMatch models.TransactionItem, companyNumber, requestId string) (*models.TransactionItem, error) {
			return wantPayablePenalty, nil
		}

		transaction := models.TransactionItem{PenaltyRef: "121"}
		gotPayablePenalty, err := PayablePenalty(generateParams(mockApDaoSvc, transaction))

		So(gotPayablePenalty, ShouldResemble, wantPayablePenalty)
		So(err, ShouldBeNil)
	})
}
