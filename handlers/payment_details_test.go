package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func serveGetPaymentDetailsHandler(payableResource *models.PayableResource) *httptest.ResponseRecorder {
	path := "/company/12345/penalties/payable/321"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	res := httptest.NewRecorder()

	if payableResource != nil {
		ctx := context.WithValue(req.Context(), config.PayableResource, payableResource)
		req = req.WithContext(ctx)
	}

	penaltyDetailsMap := &config.PenaltyDetailsMap{}
	HandleGetPaymentDetails(penaltyDetailsMap).ServeHTTP(res, req)

	return res
}

func setGetCompanyCodeFromTransactionMock(companyCode string) {
	mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return companyCode, nil
	}
	getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction
}

func TestUnitHandleGetPaymentDetails(t *testing.T) {
	Convey("No payable resource in request context", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.LateFilingPenalty)
		res := serveGetPaymentDetailsHandler(nil)
		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Payment PenaltyDetails not found due to no costs", t, func() {
		setGetCompanyCodeFromTransactionMock(utils.Sanctions)
		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CustomerCode: "12345678",
			PayableRef:   "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/abcdef",
				Payment: "/company/12345678/penalties/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		res := serveGetPaymentDetailsHandler(&payable)
		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Payment PenaltyDetails success", t, func() {
		testCases := []struct {
			name        string
			companyCode string
			penaltyRef  string
		}{
			{
				name:        "Late Filing",
				companyCode: utils.LateFilingPenalty,
				penaltyRef:  "A1234567",
			},
			{
				name:        "Sanctions",
				companyCode: utils.Sanctions,
				penaltyRef:  "P1234567",
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				setGetCompanyCodeFromTransactionMock(tc.companyCode)
				t := time.Now().Truncate(time.Millisecond)

				payable := models.PayableResource{
					CustomerCode: "12345678",
					PayableRef:   "abcdef",
					Links: models.PayableResourceLinks{
						Self:    "/company/12345678/penalties/abcdef",
						Payment: "/company/12345678/penalties/abcdef/payment",
					},
					Etag:      "qwertyetag1234",
					CreatedAt: &t,
					CreatedBy: models.CreatedBy{
						Email: "test@user.com",
						ID:    "uz3r1D_H3r3",
					},
					Transactions: []models.TransactionItem{
						{
							Amount:     5,
							Type:       "penalty",
							PenaltyRef: tc.penaltyRef,
						},
					},
					Payment: models.Payment{
						Amount: "5",
						Status: "pending",
					},
				}

				res := serveGetPaymentDetailsHandler(&payable)
				So(res.Code, ShouldEqual, http.StatusOK)
			})
		}
	})
}
