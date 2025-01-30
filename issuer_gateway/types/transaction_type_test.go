package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitTransactionType(t *testing.T) {
	Convey("Given the defined TransactionTypes", t, func() {
		testCases := []struct {
			input    TransactionType
			expected string
		}{
			{input: Penalty, expected: "penalty"},
			{input: Other, expected: "other"},
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
