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

		// check for dups by creating a map of string->int and counting the entry values whilst
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

func TestUnitGetCustomerCode(t *testing.T) {
	Convey("Get customer code", t, func() {
		testCases := []struct {
			name                 string
			input                map[string]string
			expectedCustomerCode string
			expectedError        bool
		}{
			{
				name:                 "Successful customer code",
				input:                map[string]string{"customer_code": "NI123546"},
				expectedCustomerCode: "NI123546",
				expectedError:        false,
			},
			{
				name:                 "Successful lower case customer code",
				input:                map[string]string{"customer_code": "nI123546"},
				expectedCustomerCode: "NI123546",
				expectedError:        false,
			},
			{
				name:                 "Empty customer code",
				input:                map[string]string{},
				expectedCustomerCode: "",
				expectedError:        true,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				companyNumber, err := GetCustomerCodeFromVars(tc.input)
				Convey(tc.expectedCustomerCode, func() {
					if tc.expectedError {
						So(err, ShouldNotBeNil)
					} else {
						So(err, ShouldBeNil)
					}
					So(companyNumber, ShouldEqual, tc.expectedCustomerCode)
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
				input:         LateFilingPenaltyRefType,
				expectedCode:  LateFilingPenaltyCompanyCode,
				expectedError: false,
			},
			{
				name:         "Sanctions",
				input:        SanctionsPenaltyRefType,
				expectedCode: SanctionsCompanyCode,
			},
			{
				name:         "Sanctions ROE",
				input:        SanctionsRoePenaltyRefType,
				expectedCode: SanctionsCompanyCode,
			},
			{
				name:          "Error invalid penalty reference",
				input:         "R1234567",
				expectedCode:  "",
				expectedError: true,
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
	Convey("Get Company Code from penalty ref", t, func() {
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
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "A1000007",
					},
				},
				expectedCode:  LateFilingPenaltyCompanyCode,
				expectedError: false,
			},
			{
				name: "Sanctions",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "P1000007",
					},
				},
				expectedCode: "C1",
			},
			{
				name: "Sanctions ROE",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "U1000007",
					},
				},
				expectedCode: "C1",
			},
			{
				name: "Error unknown penalty reference",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "Q1000007",
					},
				},
				expectedCode:  "",
				expectedError: true,
			},
			{
				name: "Error no penalty reference",
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

func TestUnitGetPenaltyRefTypeFromTransaction(t *testing.T) {
	Convey("Get Company Code from penalty ref", t, func() {
		testCases := []struct {
			name                   string
			input                  []models.TransactionItem
			expectedPenaltyRefType string
			expectedError          bool
		}{
			{
				name: "Late Filing",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "A1000007",
					},
				},
				expectedPenaltyRefType: LateFilingPenaltyRefType,
			},
			{
				name: "Sanctions",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "P1000007",
					},
				},
				expectedPenaltyRefType: SanctionsPenaltyRefType,
			},
			{
				name: "Sanctions ROE",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "U1000007",
					},
				},
				expectedPenaltyRefType: SanctionsRoePenaltyRefType,
			},
			{
				name: "Error unknown penalty reference",
				input: []models.TransactionItem{
					{
						Amount:     5,
						Type:       "penalty",
						PenaltyRef: "Q1000007",
					},
				},
				expectedPenaltyRefType: "",
				expectedError:          true,
			},
			{
				name: "Error no penalty reference",
				input: []models.TransactionItem{
					{},
				},
				expectedPenaltyRefType: "",
				expectedError:          true,
			},
			{
				name:                   "Error no transaction present",
				input:                  []models.TransactionItem{},
				expectedPenaltyRefType: "",
				expectedError:          true,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				penaltyRefType, err := GetPenaltyRefTypeFromTransaction(tc.input)
				Convey(tc.expectedPenaltyRefType, func() {
					if tc.expectedError {
						So(err, ShouldNotBeNil)
					} else {
						So(err, ShouldBeNil)
					}
					So(penaltyRefType, ShouldEqual, tc.expectedPenaltyRefType)
				})
			})
		}
	})
}
