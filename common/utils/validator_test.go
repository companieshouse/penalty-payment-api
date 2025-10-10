package utils

import (
	"testing"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetValidator(t *testing.T) {
	Convey("GetValidator should not return nil", t, func() {
		validator := GetValidator()
		So(validator, ShouldNotBeNil)
	})

	Convey("successful validation should not return error", t, func() {
		validator := GetValidator()
		err := validator.Validate(generatePayableRequest(true))
		So(err, ShouldBeNil)
	})

	Convey("failed validation should return error", t, func() {
		validator := GetValidator()
		err := validator.Validate(generatePayableRequest(false))
		So(err, ShouldNotBeNil)
	})
}

func generatePayableRequest(valid bool) models.PayableRequest {
	request := models.PayableRequest{
		CustomerCode: "1234",
		CreatedBy: authentication.AuthUserDetails{
			Email:    "a.o@x.com",
			Forename: "Test",
			Surname:  "Test",
			ID:       "id123",
		},
	}

	if valid {
		request.Transactions = []models.TransactionItem{
			{
				PenaltyRef: "A1234567",
				Amount:     1000,
				Type:       "penalty",
				MadeUpDate: "time.Now().String()",
				IsPaid:     true,
				IsDCA:      false,
				Reason:     "Late filing of accounts",
			},
		}
	}

	return request
}
