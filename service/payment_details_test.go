package service

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/utils"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetPaymentDetailsFromPayableResource(t *testing.T) {

	penaltyDetailsMap, err := config.LoadPenaltyDetails("../assets/penalty_details.yml")
	if err != nil {
		log.Fatal(err)
	}

	Convey("Get payment details no transactions - invalid data", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.LateFilingPenalty)

		So(paymentDetails, ShouldBeNil)
		So(responseType, ShouldEqual, InvalidData)
		So(err, ShouldNotBeNil)

	})

	Convey("Get payment details pending state - success", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{
				{
					Amount:        5,
					Type:          "penalty",
					TransactionID: "A1234567",
				},
			},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.LateFilingPenalty)

		expectedCost := models.Cost{
			Description:             "Late Filing Penalty",
			Amount:                  "5",
			AvailablePaymentMethods: []string{"credit-card"},
			ClassOfPayment:          []string{"penalty"},
			DescriptionIdentifier:   "late-filing-penalty",
			Kind:                    "cost#cost",
			ResourceKind:            "late-filing-penalty#late-filing-penalty",
			ProductType:             "late-filing-penalty",
		}

		So(paymentDetails, ShouldNotBeNil)
		So(paymentDetails.Description, ShouldEqual, "Late Filing Penalty")
		So(paymentDetails.Kind, ShouldEqual, "payment-details#payment-details")
		So(paymentDetails.PaymentReference, ShouldEqual, "")
		So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef/payment")
		So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef")
		So(paymentDetails.Status, ShouldEqual, "pending")
		So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
		So(paymentDetails.Items[0], ShouldResemble, expectedCost)
		So(responseType, ShouldEqual, Success)
		So(err, ShouldBeNil)

	})

	Convey("Get payment details paid state - success", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{
				{
					Amount:        5,
					Type:          "penalty",
					TransactionID: "0987654321",
				},
			},
			Payment: models.Payment{
				Amount:    "50",
				Status:    "paid",
				PaidAt:    &t,
				Reference: "payment_id",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.LateFilingPenalty)

		expectedCost := models.Cost{
			Description:             "Late Filing Penalty",
			Amount:                  "5",
			AvailablePaymentMethods: []string{"credit-card"},
			ClassOfPayment:          []string{"penalty"},
			DescriptionIdentifier:   "late-filing-penalty",
			Kind:                    "cost#cost",
			ResourceKind:            "late-filing-penalty#late-filing-penalty",
			ProductType:             "late-filing-penalty",
		}

		So(paymentDetails, ShouldNotBeNil)
		So(paymentDetails.Description, ShouldEqual, "Late Filing Penalty")
		So(paymentDetails.Kind, ShouldEqual, "payment-details#payment-details")
		So(paymentDetails.PaidAt, ShouldEqual, &t)
		So(paymentDetails.PaymentReference, ShouldEqual, "payment_id")
		So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef/payment")
		So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef")
		So(paymentDetails.Status, ShouldEqual, "paid")
		So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
		So(paymentDetails.Items[0], ShouldResemble, expectedCost)
		So(responseType, ShouldEqual, Success)
		So(err, ShouldBeNil)

	})
}

func TestUnitGetPaymentDetailsFromPayableResourceSanctions(t *testing.T) {

	penaltyDetailsMap, err := config.LoadPenaltyDetails("../assets/penalty_details.yml")
	if err != nil {
		log.Fatal(err)
	}

	Convey("Get payment details no transactions penalty-sanctions - invalid data", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.Sanctions)

		So(paymentDetails, ShouldBeNil)
		So(responseType, ShouldEqual, InvalidData)
		So(err, ShouldNotBeNil)

	})

	Convey("Get payment details pending state penalty-sanctions - success", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{
				{
					Amount:        5,
					Type:          "penalty",
					TransactionID: "A1234567",
				},
			},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.Sanctions)

		expectedCost := models.Cost{
			Description:             "Sanctions Penalty Payment",
			Amount:                  "5",
			AvailablePaymentMethods: []string{"credit-card"},
			ClassOfPayment:          []string{"penalty-sanctions"},
			DescriptionIdentifier:   "penalty-sanctions",
			Kind:                    "cost#cost",
			ResourceKind:            "penalty#sanctions",
			ProductType:             "penalty-sanctions",
		}

		So(paymentDetails, ShouldNotBeNil)
		So(paymentDetails.Description, ShouldEqual, "Sanctions Penalty Payment")
		So(paymentDetails.Kind, ShouldEqual, "payment-details#payment-details")
		So(paymentDetails.PaymentReference, ShouldEqual, "")
		So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef/payment")
		So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef")
		So(paymentDetails.Status, ShouldEqual, "pending")
		So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
		So(paymentDetails.Items[0], ShouldResemble, expectedCost)
		So(responseType, ShouldEqual, Success)
		So(err, ShouldBeNil)

	})

	Convey("Get payment details paid state penalty-sanctions - success", t, func() {

		path := "/company/12345678/penalties/late-filing/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

		t := time.Now().Truncate(time.Millisecond)

		payable := models.PayableResource{
			CompanyNumber: "12345678",
			Reference:     "abcdef",
			Links: models.PayableResourceLinks{
				Self:    "/company/12345678/penalties/late-filing/abcdef",
				Payment: "/company/12345678/penalties/late-filing/abcdef/payment",
			},
			Etag:      "qwertyetag1234",
			CreatedAt: &t,
			CreatedBy: models.CreatedBy{
				Email: "test@user.com",
				ID:    "uz3r1D_H3r3",
			},
			Transactions: []models.TransactionItem{
				{
					Amount:        5,
					Type:          "penalty",
					TransactionID: "0987654321",
				},
			},
			Payment: models.Payment{
				Amount:    "50",
				Status:    "paid",
				PaidAt:    &t,
				Reference: "payment_id",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, responseType, err := service.GetPaymentDetailsFromPayableResource(req, &payable,
			penaltyDetailsMap, utils.Sanctions)

		expectedCost := models.Cost{
			Description:             "Sanctions Penalty Payment",
			Amount:                  "5",
			AvailablePaymentMethods: []string{"credit-card"},
			ClassOfPayment:          []string{"penalty-sanctions"},
			DescriptionIdentifier:   "penalty-sanctions",
			Kind:                    "cost#cost",
			ResourceKind:            "penalty#sanctions",
			ProductType:             "penalty-sanctions",
		}

		So(paymentDetails, ShouldNotBeNil)
		So(paymentDetails.Description, ShouldEqual, "Sanctions Penalty Payment")
		So(paymentDetails.Kind, ShouldEqual, "payment-details#payment-details")
		So(paymentDetails.PaidAt, ShouldEqual, &t)
		So(paymentDetails.PaymentReference, ShouldEqual, "payment_id")
		So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef/payment")
		So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/late-filing/abcdef")
		So(paymentDetails.Status, ShouldEqual, "paid")
		So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
		So(paymentDetails.Items[0], ShouldResemble, expectedCost)
		So(responseType, ShouldEqual, Success)
		So(err, ShouldBeNil)

	})
}
