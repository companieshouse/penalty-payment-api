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

func TestUnitPayablePenalty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccountPenaltiesDaoService := mocks.NewMockAccountPenaltiesDaoService(ctrl)

	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionMap := &models.AllowedTransactionMap{
		Types: map[string]map[string]bool{
			"1": {
				"EJ": true,
				"EU": true,
			},
		},
	}

	Convey("error is returned when fetching account penalties fails", t, func() {
		accountPenaltiesErr := errors.New("failed to fetch account penalties")
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
			return nil, services.Error, accountPenaltiesErr
		}
		getAccountPenalties = mockedAccountPenalties

		transaction := models.TransactionItem{PenaltyRef: "121"}
		payablePenalty, err := PayablePenalty("10000024", utils.LateFilingPenalty, transaction,
			penaltyDetailsMap, allowedTransactionMap, mockAccountPenaltiesDaoService)

		So(payablePenalty, ShouldBeNil)
		So(err, ShouldEqual, accountPenaltiesErr)
	})

	Convey("payable penalty is successfully returned for multiple unpaid penalties", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
			return accountPenaltiesResponse(2), services.Success, nil
		}
		getAccountPenalties = mockedAccountPenalties

		unpaidPenalty := models.TransactionItem{
			PenaltyRef: "A0000002",
			Amount:     150,
			Type:       "penalty",
			MadeUpDate: "2017-06-30",
			IsDCA:      false,
			IsPaid:     false,
			Reason:     "Late filing of accounts",
		}
		gotPayablePenalty, err := PayablePenalty("10000024", utils.LateFilingPenalty, unpaidPenalty, penaltyDetailsMap,
			allowedTransactionMap, mockAccountPenaltiesDaoService)

		So(gotPayablePenalty, ShouldResemble, &unpaidPenalty)
		So(err, ShouldBeNil)
	})

	Convey("payable penalty is successfully returned", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
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

		transaction := models.TransactionItem{PenaltyRef: "121"}
		mockedMatchPenalty := func(referenceTransactions []models.TransactionListItem, transactionToMatch models.TransactionItem, companyNumber string) (*models.TransactionItem, error) {
			return wantPayablePenalty, nil
		}
		getAccountPenalties = mockedAccountPenalties
		getMatchingPenalty = mockedMatchPenalty

		gotPayablePenalty, err := PayablePenalty("10000024", utils.LateFilingPenalty, transaction,
			penaltyDetailsMap, allowedTransactionMap, mockAccountPenaltiesDaoService)

		So(gotPayablePenalty, ShouldResemble, wantPayablePenalty)
		So(err, ShouldBeNil)
	})
}
