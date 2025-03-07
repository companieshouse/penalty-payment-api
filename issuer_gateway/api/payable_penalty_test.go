package api

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
	. "github.com/smartystreets/goconvey/convey"
)

func accountPenaltiesResponse(unpaidPenaltyCount int) *models.TransactionListResponse {
	id := "12"
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
	}

	paidPenalty := unpaidPenalty
	paidPenalty.ID = "00482775"
	paidPenalty.IsPaid = true

	unpaidOther := unpaidPenalty
	unpaidOther.ID = "00482776"
	unpaidOther.Type = "other"

	response := models.TransactionListResponse{
		Items: []models.TransactionListItem{paidPenalty, unpaidOther},
	}

	for i := 0; i < unpaidPenaltyCount; i++ {
		item := unpaidPenalty
		item.ID = id + string(rune(i))
		response.Items = append(response.Items, item)
	}

	return &response
}

func TestUnitPayablePenalty(t *testing.T) {

	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	allowedTransactionMap := &models.AllowedTransactionMap{
		Types: map[string]map[string]bool{
			"1": {
				"EJ": true,
				"EU": true,
			},
		},
	}
	transaction := models.TransactionItem{TransactionID: "121"}

	Convey("error is returned when fetching account penalties fails", t, func() {
		accountPenaltiesErr := errors.New("failed to fetch account penalties")
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
			return nil, services.Error, accountPenaltiesErr
		}
		getAccountPenalties = mockedAccountPenalties

		payablePenalty, err := PayablePenalty("10000024", "LP", transaction, penaltyDetailsMap, allowedTransactionMap)

		So(payablePenalty, ShouldBeNil)
		So(err, ShouldEqual, accountPenaltiesErr)
	})

	Convey("error is returned for multiple unpaid penalties", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
			return accountPenaltiesResponse(2), services.Success, nil
		}
		getAccountPenalties = mockedAccountPenalties

		_, err := PayablePenalty("10000024", "LP", transaction, penaltyDetailsMap, allowedTransactionMap)

		So(err, ShouldBeError, private.ErrMultiplePenalties)
	})

	Convey("payable penalty is successfully returned", t, func() {
		mockedAccountPenalties := func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
			return accountPenaltiesResponse(1), services.Success, nil
		}
		wantPayablePenalty := &models.TransactionItem{
			TransactionID: "121",
			Amount:        150,
			Type:          "penalty",
			MadeUpDate:    "2017-06-30",
			IsDCA:         false,
			IsPaid:        false,
		}
		mockedMatchPenalty := func(referenceTransactions []models.TransactionListItem, transactionToMatch models.TransactionItem, companyNumber string) (*models.TransactionItem, error) {
			return wantPayablePenalty, nil
		}
		getAccountPenalties = mockedAccountPenalties
		getMatchingPenalty = mockedMatchPenalty

		gotPayablePenalty, err := PayablePenalty("10000024", "LP", transaction, penaltyDetailsMap, allowedTransactionMap)

		So(gotPayablePenalty, ShouldResemble, wantPayablePenalty)
		So(err, ShouldBeNil)
	})
}
