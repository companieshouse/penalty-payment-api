package private

import (
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitMatchPenalty(t *testing.T) {

	companyNumber := "123"
	transactionsToMatch := []models.TransactionItem{
		{
			TransactionID: "121",
			Type:          "penalty",
			Amount:        150,
			Reason:        "Failure to file a confirmation statement",
		},
	}
	matchedPenalties := []models.TransactionItem{
		{
			TransactionID: "121",
			Amount:        150,
			Type:          "penalty",
			MadeUpDate:    "2017-06-30",
			Reason:        "Failure to file a confirmation statement",
			IsDCA:         false,
			IsPaid:        false,
		},
	}

	testCases := []struct {
		TransactionID  string
		Type           string
		MadeUpDate     string
		Reason         string
		IsDCA          bool
		IsPaid         bool
		OriginalAmount float64
		Outstanding    float64
		WantMatched    []models.TransactionItem
		WantError      error
	}{
		{TransactionID: "120", Outstanding: 150, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 150, IsDCA: false, IsPaid: false, WantMatched: nil, WantError: ErrTransactionDoesNotExist},
		{TransactionID: "121", Outstanding: 150, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 200, IsDCA: false, IsPaid: false, WantMatched: nil, WantError: ErrTransactionIsPartPaid},
		{TransactionID: "121", Outstanding: 150, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 150, IsDCA: false, IsPaid: true, WantMatched: nil, WantError: ErrTransactionIsPaid},
		{TransactionID: "121", Outstanding: 100, Type: "other", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 100, IsDCA: false, IsPaid: false, WantMatched: nil, WantError: ErrTransactionNotPayable},
		{TransactionID: "121", Outstanding: 100, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 100, IsDCA: false, IsPaid: false, WantMatched: nil, WantError: ErrTransactionAmountMismatch},
		{TransactionID: "121", Outstanding: 150, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 150, IsDCA: true, IsPaid: false, WantMatched: nil, WantError: ErrTransactionDCA},
		{TransactionID: "121", Outstanding: 150, Type: "penalty", MadeUpDate: "2017-06-30", Reason: "Failure to file a confirmation statement",
			OriginalAmount: 150, IsDCA: false, IsPaid: false, WantMatched: matchedPenalties, WantError: nil},
	}

	Convey("matchPenalty works correctly for different scenarios", t, func() {
		for _, testCase := range testCases {
			refTransactions := []models.TransactionListItem{
				{
					ID:             testCase.TransactionID,
					Type:           testCase.Type,
					OriginalAmount: testCase.OriginalAmount,
					Outstanding:    testCase.Outstanding,
					IsDCA:          testCase.IsDCA,
					IsPaid:         testCase.IsPaid,
					MadeUpDate:     testCase.MadeUpDate,
					Reason:         testCase.Reason,
				},
			}
			matched, err := MatchPenalty(refTransactions, transactionsToMatch, companyNumber)

			So(err, ShouldEqual, testCase.WantError)
			So(matched, ShouldResemble, testCase.WantMatched)
		}
	})
}
