package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitResponseType(t *testing.T) {
	Convey("Given String is called on a ResponseType", t, func() {
		Convey("When a valid value is provided", func() {
			val := ResponseType(0)
			result := val.String()

			Convey("Then the correct s should be returned", func() {
				So(result, ShouldEqual, "Success")
			})
		})

		Convey("When an invalid value is provided", func() {
			val := ResponseType(123)
			result := val.String()

			Convey("Then 'invalid-data' should be returned", func() {
				So(result, ShouldEqual, "invalid-data")
			})
		})
	})
}
