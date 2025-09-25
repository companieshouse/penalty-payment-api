package utils

import (
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGetValidator(t *testing.T) {
	Convey("Validator with empty input", t, func() {
		validator := GetValidator(t)
		So(validator, ShouldBeNil)
	})

	Convey("Validator with empty input", t, func() {
		var request models.PayableRequest
		validator := GetValidator(request)
		So(validator, ShouldNotBeNil)
	})
}
