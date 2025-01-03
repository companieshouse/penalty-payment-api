package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitWriteJSON(t *testing.T) {
	Convey("Failure to marshal json", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		// causes an UnsupportedTypeError
		WriteJSON(w, r, make(chan int))

		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		So(w.Body.String(), ShouldEqual, "")
	})

	Convey("contents are written as json", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		WriteJSON(w, r, &models.CreatedPayableResourceLinks{})

		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		So(w.Body.String(), ShouldEqual, "{\"self\":\"\"}\n")
	})
}

func TestUnitGetCompanyNumber(t *testing.T) {
	Convey("Given GetCompanyNumber is called", t, func() {
		companyNumber := "NI12356"
		Convey("When a company number is provided", func() {
			vars := map[string]string{
				"company_number": companyNumber,
			}
			result, err := GetCompanyNumberFromVars(vars)

			Convey("Then err should be nil and company number should equal "+companyNumber, func() {
				So(result, ShouldEqual, companyNumber)
				So(err, ShouldBeNil)
			})
		})
		Convey("When no company number is provided", func() {
			vars := map[string]string{}
			result, err := GetCompanyNumberFromVars(vars)
			Convey("Then err should be thrown", func() {
				So(result, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, "company number not supplied")
			})
		})
	})
}

// this function always return "LP" at the moment
func TestUnitGetCompanyCodeFromVars(t *testing.T) {
	Convey("Get Company Code from vars", t, func() {
		companyNumber, err := GetCompanyCodeFromVars()
		So(companyNumber, ShouldEqual, "LP")
		So(err, ShouldBeNil)
	})
}
