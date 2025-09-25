package handlers

import (
	"context"
	"errors"
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

func setGetPenaltyRefTypeFromTransactionMock(penaltyRefType string) {
	mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return penaltyRefType, nil
	}
	getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction
}

func generateTestPayableResource(withTransaction bool, penaltyRef string) models.PayableResource {
	t := time.Now().Truncate(time.Millisecond)
	p := models.PayableResource{
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
	if withTransaction {
		p.Transactions = []models.TransactionItem{
			{
				Amount:     5,
				Type:       "penalty",
				PenaltyRef: penaltyRef,
			},
		}
	}
	return p
}

func TestUnitHandleGetPaymentDetails(t *testing.T) {
	Convey("No payable resource in request context", t, func() {
		setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenRef)

		res := serveGetPaymentDetailsHandler(nil)
		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Cannot determine penalty ref type from transaction ID", t, func() {

		getPenaltyRefTypeFromTransaction = func(transactions []models.TransactionItem) (string, error) {
			return "", errors.New("cannot determine penalty ref type")
		}

		payable := generateTestPayableResource(false, "")

		res := serveGetPaymentDetailsHandler(&payable)
		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Payment PenaltyDetails not found due to no costs", t, func() {
		setGetPenaltyRefTypeFromTransactionMock(utils.SanctionsPenRef)

		payable := generateTestPayableResource(false, "")

		res := serveGetPaymentDetailsHandler(&payable)
		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Payment PenaltyDetails success", t, func() {
		testCases := []struct {
			name           string
			companyCode    string
			penaltyRefType string
			penaltyRef     string
		}{
			{
				name:           "Late Filing",
				companyCode:    utils.LateFilingPenaltyCompanyCode,
				penaltyRefType: utils.LateFilingPenRef,
				penaltyRef:     "A1234567",
			},
			{
				name:           "Sanctions",
				companyCode:    utils.SanctionsCompanyCode,
				penaltyRefType: utils.SanctionsPenRef,
				penaltyRef:     "P1234567",
			},
			{
				name:           "Sanctions ROE",
				companyCode:    utils.SanctionsCompanyCode,
				penaltyRefType: utils.SanctionsRoePenRef,
				penaltyRef:     "U1234567",
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				setGetPenaltyRefTypeFromTransactionMock(tc.penaltyRefType)

				payable := generateTestPayableResource(true, tc.penaltyRef)

				res := serveGetPaymentDetailsHandler(&payable)
				So(res.Code, ShouldEqual, http.StatusOK)
			})
		}
	})
}
