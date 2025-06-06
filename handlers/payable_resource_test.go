package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHandleGetPayableResource(t *testing.T) {
	Convey("Invalid PayableResourceRest", t, func() {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		HandleGetPayableResource(w, req)
		So(w.Code, ShouldEqual, 500)
	})
	Convey("Valid PayableResource", t, func() {
		createdAt := time.Now().Truncate(time.Millisecond)
		payable := models.PayableResource{
			CustomerCode: "12345678",
			PayableRef:   "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/abcdef",
				Payment: "/company/12345678/penalties/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &createdAt,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{
				{
					Amount:     5,
					Type:       "penalty",
					PenaltyRef: "A1234567",
				},
			},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), config.PayableResource, &payable)
		w := httptest.NewRecorder()

		HandleGetPayableResource(w, req.WithContext(ctx))

		So(w.Code, ShouldEqual, 200)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

		resultPayable := &models.PayableResource{}
		json.NewDecoder(w.Body).Decode(&resultPayable)
		So(resultPayable.CustomerCode, ShouldEqual, payable.CustomerCode)
		So(resultPayable.PayableRef, ShouldEqual, payable.PayableRef)
		So(resultPayable.Etag, ShouldEqual, payable.Etag)
		So(resultPayable.CreatedAt.Nanosecond(), ShouldEqual, payable.CreatedAt.Nanosecond())
		So(resultPayable.Links.Self, ShouldEqual, payable.Links.Self)
		So(resultPayable.Links.Payment, ShouldEqual, payable.Links.Payment)
		So(resultPayable.CreatedBy.Email, ShouldEqual, payable.CreatedBy.Email)
		So(resultPayable.CreatedBy.ID, ShouldEqual, payable.CreatedBy.ID)
		So(resultPayable.Payment.Amount, ShouldEqual, payable.Payment.Amount)
		So(resultPayable.Payment.Status, ShouldEqual, payable.Payment.Status)
		So(len(resultPayable.Transactions), ShouldEqual, 1)
		So(resultPayable.Transactions[0].Amount, ShouldEqual, payable.Transactions[0].Amount)
		So(resultPayable.Transactions[0].Type, ShouldEqual, payable.Transactions[0].Type)
		So(resultPayable.Transactions[0].PenaltyRef, ShouldEqual, payable.Transactions[0].PenaltyRef)

	})
}
