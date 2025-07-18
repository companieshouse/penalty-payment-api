package service

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetPaymentDetailsFromPayableResource(t *testing.T) {

	penaltyDetailsMap, err := config.LoadPenaltyDetails("../../assets/penalty_details.yml")
	if err != nil {
		log.Fatal(err)
	}
	penaltyDetails := penaltyDetailsMap.Details[utils.LateFilingPenaltyCompanyCode]

	Convey("Get payment details no transactions - invalid data", t, func() {

		path := "/company/12345678/penalties/abcdef/payment"
		req := httptest.NewRequest(http.MethodGet, path, nil)

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
			Transactions: []models.TransactionItem{},
			Payment: models.Payment{
				Amount: "5",
				Status: "pending",
			},
		}

		service := &PaymentDetailsService{}

		paymentDetails, err := service.GetPaymentDetailsFromPayableResource(req, &payable, penaltyDetails)

		So(paymentDetails, ShouldBeNil)
		So(err, ShouldNotBeNil)

	})

	Convey("Get payment details pending state - success", t, func() {

		testCases := []struct {
			description           string
			kind                  string
			classOfPayment        string
			descriptionIdentifier string
			resourceKind          string
			productType           string
			companyCode           string
			penaltyRefType        string
		}{
			{
				description:           "Late Filing Penalty",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-lfp",
				descriptionIdentifier: "late-filing-penalty",
				resourceKind:          "late-filing-penalty#late-filing-penalty",
				productType:           "late-filing-penalty",
				companyCode:           utils.LateFilingPenaltyCompanyCode,
				penaltyRefType:        utils.LateFilingPenRef,
			},
			{
				description:           "Sanctions Penalty Payment",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-sanctions",
				descriptionIdentifier: "penalty-sanctions",
				resourceKind:          "penalty#sanctions",
				productType:           "penalty-sanctions",
				companyCode:           utils.SanctionsCompanyCode,
				penaltyRefType:        utils.SanctionsPenRef,
			},
			{
				description:           "Overseas Entity Penalty Payment",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-sanctions",
				descriptionIdentifier: "penalty-sanctions",
				resourceKind:          "penalty#sanctions",
				productType:           "penalty-sanctions",
				companyCode:           utils.SanctionsCompanyCode,
				penaltyRefType:        utils.SanctionsRoePenRef,
			},
		}
		for _, tc := range testCases {
			Convey(tc.penaltyRefType, func() {
				path := "/company/12345678/penalties/abcdef/payment"
				req := httptest.NewRequest(http.MethodGet, path, nil)

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
							PenaltyRef: "A1234567",
						},
					},
					Payment: models.Payment{
						Amount: "5",
						Status: "pending",
					},
				}

				service := &PaymentDetailsService{}

				penaltyDetails := penaltyDetailsMap.Details[tc.penaltyRefType]
				paymentDetails, err := service.GetPaymentDetailsFromPayableResource(req, &payable, penaltyDetails)

				expectedCost := models.Cost{
					Description:             tc.description,
					Amount:                  "5",
					AvailablePaymentMethods: []string{"credit-card"},
					ClassOfPayment:          []string{tc.classOfPayment},
					DescriptionIdentifier:   tc.descriptionIdentifier,
					Kind:                    "cost#cost",
					ResourceKind:            tc.resourceKind,
					ProductType:             tc.productType,
				}

				So(paymentDetails, ShouldNotBeNil)
				So(paymentDetails.Description, ShouldEqual, tc.description)
				So(paymentDetails.Kind, ShouldEqual, tc.resourceKind)
				So(paymentDetails.PaymentReference, ShouldEqual, "")
				So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/abcdef/payment")
				So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/abcdef")
				So(paymentDetails.Status, ShouldEqual, "pending")
				So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
				So(paymentDetails.Items[0], ShouldResemble, expectedCost)
				So(err, ShouldBeNil)
			})
		}

	})

	Convey("Get payment details paid state - success", t, func() {

		testCases := []struct {
			description           string
			kind                  string
			classOfPayment        string
			descriptionIdentifier string
			resourceKind          string
			productType           string
			companyCode           string
			penaltyRefType        string
		}{
			{
				description:           "Late Filing Penalty",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-lfp",
				descriptionIdentifier: "late-filing-penalty",
				resourceKind:          "late-filing-penalty#late-filing-penalty",
				productType:           "late-filing-penalty",
				companyCode:           utils.LateFilingPenaltyCompanyCode,
				penaltyRefType:        utils.LateFilingPenRef,
			},
			{
				description:           "Sanctions Penalty Payment",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-sanctions",
				descriptionIdentifier: "penalty-sanctions",
				resourceKind:          "penalty#sanctions",
				productType:           "penalty-sanctions",
				companyCode:           utils.SanctionsCompanyCode,
				penaltyRefType:        utils.SanctionsPenRef,
			},
			{
				description:           "Overseas Entity Penalty Payment",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-sanctions",
				descriptionIdentifier: "penalty-sanctions",
				resourceKind:          "penalty#sanctions",
				productType:           "penalty-sanctions",
				companyCode:           utils.SanctionsCompanyCode,
				penaltyRefType:        utils.SanctionsRoePenRef,
			},
		}
		for _, tc := range testCases {
			Convey(tc.penaltyRefType, func() {
				path := "/company/12345678/penalties/abcdef/payment"
				req := httptest.NewRequest(http.MethodGet, path, nil)

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
							PenaltyRef: "0987654321",
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

				penaltyDetails := penaltyDetailsMap.Details[tc.penaltyRefType]
				paymentDetails, err := service.GetPaymentDetailsFromPayableResource(req, &payable, penaltyDetails)

				expectedCost := models.Cost{
					Description:             tc.description,
					Amount:                  "5",
					AvailablePaymentMethods: []string{"credit-card"},
					ClassOfPayment:          []string{tc.classOfPayment},
					DescriptionIdentifier:   tc.descriptionIdentifier,
					Kind:                    "cost#cost",
					ResourceKind:            tc.resourceKind,
					ProductType:             tc.productType,
				}

				So(paymentDetails, ShouldNotBeNil)
				So(paymentDetails.Description, ShouldEqual, tc.description)
				So(paymentDetails.Kind, ShouldEqual, tc.resourceKind)
				So(paymentDetails.PaidAt, ShouldEqual, &t)
				So(paymentDetails.PaymentReference, ShouldEqual, "payment_id")
				So(paymentDetails.Links.Self, ShouldEqual, "/company/12345678/penalties/abcdef/payment")
				So(paymentDetails.Links.Resource, ShouldEqual, "/company/12345678/penalties/abcdef")
				So(paymentDetails.Status, ShouldEqual, "paid")
				So(paymentDetails.CompanyNumber, ShouldEqual, "12345678")
				So(paymentDetails.Items[0], ShouldResemble, expectedCost)
				So(err, ShouldBeNil)
			})
		}
	})
}
