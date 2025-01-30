package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitResponseType(t *testing.T) {
	Convey("Given the defined ResponseTypes", t, func() {
		testCases := []struct {
			input    ResponseType
			expected string
		}{
			{input: InvalidData, expected: "invalid-data"},
			{input: Error, expected: "error"},
			{input: Forbidden, expected: "forbidden"},
			{input: NotFound, expected: "not-found"},
			{input: Success, expected: "success"},
		}
		Convey("When String is called", func() {
			for _, testCase := range testCases {
				testCase := testCase

				Convey("Then the correct string should be returned for "+testCase.expected, func() {
					result := testCase.input.String()
					So(result, ShouldEqual, testCase.expected)
				})
			}
		})
	})
}
