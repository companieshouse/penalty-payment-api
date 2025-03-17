package utils

import (
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitGenerateReferenceNumber(t *testing.T) {
	Convey("Reference Number is correct length", t, func() {
		ref := GenerateReferenceNumber()
		So(len(ref), ShouldEqual, 10)
	})

	Convey("Reference Number does not collide", t, func() {
		// generate 10,000 reference numbers and check for any duplicates
		times := 10000 // 10 thousand
		generated := make([]string, times)

		for i := 0; i < times; i++ {
			ref := GenerateReferenceNumber()
			generated[i] = ref
		}

		// check for dups by creating a map of string->int and counting the the entry values whilst
		// iterating through the generated map
		generatedCheck := make(map[string]int)
		for _, reference := range generated {
			_, exists := generatedCheck[reference]
			So(exists, ShouldBeZeroValue)
			generatedCheck[reference] = 1
		}
	})
}

func TestUnitGenerateEtag(t *testing.T) {
	Convey("Generate Etag", t, func() {
		etag, err := GenerateEtag()
		So(len(etag), ShouldEqual, 56)
		So(err, ShouldBeNil)
		So(err, ShouldBeNil)
	})
}

func TestUnitGetCompanyNumber(t *testing.T) {
	Convey("Get Company Number", t, func() {
		testCases := []struct {
			name                  string
			input                 map[string]string
			expectedCompanyNumber string
			expectedError         bool
		}{
			{
				name:                  "Successful company number",
				input:                 map[string]string{"company_number": "NI123546"},
				expectedCompanyNumber: "NI123546",
				expectedError:         false,
			},
			{
				name:                  "Successful lower case company number",
				input:                 map[string]string{"company_number": "nI123546"},
				expectedCompanyNumber: "NI123546",
				expectedError:         false,
			},
			{
				name:                  "Empty company number",
				input:                 map[string]string{},
				expectedCompanyNumber: "",
				expectedError:         true,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				companyNumber, err := GetCompanyNumberFromVars(tc.input)
				Convey(tc.expectedCompanyNumber, func() {
					if tc.expectedError {
						So(err, ShouldNotBeNil)
					} else {
						So(err, ShouldBeNil)
					}
					So(companyNumber, ShouldEqual, tc.expectedCompanyNumber)
				})
			})
		}
	})
}

func TestUnitGetCompanyCode(t *testing.T) {
	Convey("Get Company Code from penalty reference", t, func() {
		testCases := []struct {
			name          string
			input         string
			expectedCode  string
			expectedError bool
		}{
			{
				name:          "Late Filing",
				input:         "LATE_FILING",
				expectedCode:  LateFilingPenalty,
				expectedError: false,
			},
			{
				name:         "Sanctions",
				input:        "SANCTIONS",
				expectedCode: Sanctions,
			},
			{
				name:          "Error invalid penalty reference",
				input:         "R1234567",
				expectedCode:  "",
				expectedError: true,
			},
			{
				name:          "No penalty reference - default to LFP",
				input:         "",
				expectedCode:  LateFilingPenalty,
				expectedError: false,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				companyCode, err := GetCompanyCode(tc.input)
				Convey(tc.expectedCode, func() {
					if tc.expectedError {
						So(err, ShouldNotBeNil)
					} else {
						So(err, ShouldBeNil)
					}
					So(companyCode, ShouldEqual, tc.expectedCode)
				})
			})
		}
	})
}

func TestUnitGetCompanyCodeFromTransaction(t *testing.T) {
	Convey("Get Company Code from transaction ID", t, func() {
		testCases := []struct {
			name          string
			input         []models.TransactionItem
			expectedCode  string
			expectedError bool
		}{
			{
				name: "Late Filing",
				input: []models.TransactionItem{
					{
						Amount:        5,
						Type:          "penalty",
						TransactionID: "A1000007",
					},
				},
				expectedCode:  LateFilingPenalty,
				expectedError: false,
			},
			{
				name: "Sanctions",
				input: []models.TransactionItem{
					{
						Amount:        5,
						Type:          "penalty",
						TransactionID: "P1000007",
					},
				},
				expectedCode: "C1",
			},
			{
				name: "Error unknown transaction ID",
				input: []models.TransactionItem{
					{
						Amount:        5,
						Type:          "penalty",
						TransactionID: "Q1000007",
					},
				},
				expectedCode:  "",
				expectedError: true,
			},
			{
				name: "Error no transaction ID",
				input: []models.TransactionItem{
					{},
				},
				expectedCode:  "",
				expectedError: true,
			},
			{
				name:          "Error no transaction present",
				input:         []models.TransactionItem{},
				expectedCode:  "",
				expectedError: true,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				companyCode, err := GetCompanyCodeFromTransaction(tc.input)
				Convey(tc.expectedCode, func() {
					if tc.expectedError {
						So(err, ShouldNotBeNil)
					} else {
						So(err, ShouldBeNil)
					}
					So(companyCode, ShouldEqual, tc.expectedCode)
				})
			})
		}
	})
}
