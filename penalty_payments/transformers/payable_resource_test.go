package transformers

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitPayableResourceRequestToDB(t *testing.T) {
	Convey("reference number is generated", t, func() {
		req := &models.PayableRequest{
			Transactions: []models.TransactionItem{
				{TransactionID: "123"},
			},
		}
		dao := PayableResourceRequestToDB(req)

		So(dao.PayableRef, ShouldHaveLength, 10)
	})

	Convey("self link is constructed correctly", t, func() {
		req := &models.PayableRequest{
			CompanyNumber: "00006400",
			Transactions: []models.TransactionItem{
				{TransactionID: "123"},
			},
		}
		dao := PayableResourceRequestToDB(req)

		// ensure a reference is generated for the next assertion
		So(dao.PayableRef, ShouldHaveLength, 10)

		expected := fmt.Sprintf("/company/%s/financial-penalties/payable/%s", req.CompanyNumber, dao.PayableRef)
		So(dao.Data.Links.Self, ShouldContainSubstring, expected)
		So(dao.Data.Links.ResumeJourney, ShouldEqual, "/pay-penalty/company/00006400/penalty/123/view-penalties")
	})
}

func TestUnitPayableResourceDaoToCreatedResponse(t *testing.T) {
	Convey("link to self is correct", t, func() {
		dao := &models.PayableResourceDao{
			PayableRef: "1234",
			Data: models.PayableResourceDataDao{
				Links: models.PayableResourceLinksDao{
					Self:          "/foo",
					ResumeJourney: "/company/12345678/penalty/123/lfp/view-penalties",
				},
				Transactions: map[string]models.TransactionDao{
					"123": {Amount: 100, Type: "penalty", MadeUpDate: "2019-01-01"},
				},
			},
		}

		response := PayableResourceDaoToCreatedResponse(dao)

		So(response.Links.Self, ShouldEqual, dao.Data.Links.Self)
	})
}

func TestUnitPayableResourceDBToPayableResource(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		t := time.Now().Truncate(time.Millisecond)
		dao := &models.PayableResourceDao{
			CustomerCode: "12345678",
			PayableRef:   "1234",
			Data: models.PayableResourceDataDao{
				Etag:      "qwertyetag1234",
				CreatedAt: &t,
				CreatedBy: models.CreatedByDao{
					ID:       "uz3r_1d",
					Email:    "test@user.com",
					Forename: "some",
					Surname:  "body",
				},
				Links: models.PayableResourceLinksDao{
					Self:          "/foo",
					Payment:       "/foo/pay",
					ResumeJourney: "/company/12345678/penalty/123/lfp/view-penalties",
				},
				Payment: models.PaymentDao{
					Amount:    "100",
					Status:    "pending",
					Reference: "payref",
					PaidAt:    &t,
				},
				Transactions: map[string]models.TransactionDao{
					"123": {
						Amount:     100,
						Type:       "penalty",
						MadeUpDate: "2019-01-01",
					},
				},
			},
		}

		response := PayableResourceDBToRequest(dao)

		So(response.CompanyNumber, ShouldEqual, dao.CustomerCode)
		So(response.Reference, ShouldEqual, dao.PayableRef)
		So(response.Etag, ShouldEqual, dao.Data.Etag)
		So(response.CreatedAt, ShouldEqual, dao.Data.CreatedAt)
		So(response.CreatedBy.ID, ShouldEqual, dao.Data.CreatedBy.ID)
		So(response.CreatedBy.Email, ShouldEqual, dao.Data.CreatedBy.Email)
		So(response.CreatedBy.Forename, ShouldEqual, dao.Data.CreatedBy.Forename)
		So(response.CreatedBy.Surname, ShouldEqual, dao.Data.CreatedBy.Surname)
		So(response.Links.Self, ShouldEqual, dao.Data.Links.Self)
		So(response.Links.Payment, ShouldEqual, dao.Data.Links.Payment)
		So(response.Links.ResumeJourney, ShouldEqual, dao.Data.Links.ResumeJourney)
		So(response.Payment.Amount, ShouldEqual, dao.Data.Payment.Amount)
		So(response.Payment.Status, ShouldEqual, dao.Data.Payment.Status)
		So(response.Payment.Reference, ShouldEqual, dao.Data.Payment.Reference)
		So(response.Payment.PaidAt, ShouldEqual, dao.Data.Payment.PaidAt)
		So(len(response.Transactions), ShouldEqual, 1)
		So(response.Transactions[0].Amount, ShouldEqual, dao.Data.Transactions["123"].Amount)
		So(response.Transactions[0].Type, ShouldEqual, dao.Data.Transactions["123"].Type)
		So(response.Transactions[0].MadeUpDate, ShouldEqual, dao.Data.Transactions["123"].MadeUpDate)
	})
}

func TestUnitPayableResourceToPaymentDetails(t *testing.T) {
	Convey("field mappings are correct from payable resource to payment details", t, func() {
		testCases := []struct {
			description           string
			kind                  string
			classOfPayment        string
			descriptionIdentifier string
			resourceKind          string
			productType           string
			companyCode           string
		}{
			{
				description:           "Late Filing Penalty",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty",
				descriptionIdentifier: "late-filing-penalty",
				resourceKind:          "late-filing-penalty#late-filing-penalty",
				productType:           "late-filing-penalty",
				companyCode:           utils.LateFilingPenalty,
			},
			{
				description:           "Sanctions Penalty Payment",
				kind:                  "payment-details#payment-details",
				classOfPayment:        "penalty-sanctions",
				descriptionIdentifier: "penalty-sanctions",
				resourceKind:          "penalty#sanctions",
				productType:           "penalty-sanctions",
				companyCode:           utils.Sanctions,
			},
		}
		for _, tc := range testCases {
			Convey(tc.description, func() {
				t := time.Now().Truncate(time.Millisecond)
				payable := &models.PayableResource{
					CompanyNumber: "12345678",
					Reference:     "1234",
					Etag:          "qwertyetag1234",
					CreatedAt:     &t,
					CreatedBy: models.CreatedBy{
						ID:       "uz3r_1d",
						Email:    "test@user.com",
						Forename: "some",
						Surname:  "body",
					},
					Links: models.PayableResourceLinks{
						Self:    "/foo",
						Payment: "/foo/pay",
					},
					Payment: models.Payment{
						Amount:    "100",
						Status:    "pending",
						Reference: "payref",
						PaidAt:    &t,
					},
					Transactions: []models.TransactionItem{
						{
							Amount:     100,
							Type:       "penalty",
							MadeUpDate: "2019-01-01",
						},
					},
				}

				penaltyDetailsMap, err := config.LoadPenaltyDetails("../../assets/penalty_details.yml")
				if err != nil {
					log.Fatal(err)
				}
				penaltyDetails := penaltyDetailsMap.Details[tc.companyCode]

				response := PayableResourceToPaymentDetails(payable, penaltyDetails)

				_, filename, _, _ := runtime.Caller(0)
				fmt.Printf("Current test filename: %s\n", filename)

				dir, err := os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Dir: " + dir)

				So(response, ShouldNotBeNil)
				So(response.Description, ShouldEqual, tc.description)
				So(response.Kind, ShouldEqual, tc.resourceKind)
				So(response.PaidAt, ShouldEqual, payable.Payment.PaidAt)
				So(response.PaymentReference, ShouldEqual, payable.Payment.Reference)
				So(response.Links.Self, ShouldEqual, payable.Links.Payment)
				So(response.Links.Resource, ShouldEqual, payable.Links.Self)
				So(response.Status, ShouldEqual, payable.Payment.Status)
				So(response.CompanyNumber, ShouldEqual, payable.CompanyNumber)
				So(len(response.Items), ShouldEqual, 1)
				So(response.Items[0].Amount, ShouldEqual, fmt.Sprintf("%g", payable.Transactions[0].Amount))
				So(response.Items[0].AvailablePaymentMethods, ShouldResemble, []string{"credit-card"})
				So(response.Items[0].ClassOfPayment, ShouldResemble, []string{tc.classOfPayment})
				So(response.Items[0].Description, ShouldEqual, tc.description)
				So(response.Items[0].DescriptionIdentifier, ShouldEqual, tc.descriptionIdentifier)
				So(response.Items[0].Kind, ShouldEqual, "cost#cost")
				So(response.Items[0].ResourceKind, ShouldEqual, tc.resourceKind)
				So(response.Items[0].ProductType, ShouldEqual, tc.productType)
			})
		}
	})
}
