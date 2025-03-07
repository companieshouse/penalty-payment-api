package private

import (
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitLogE5Error(t *testing.T) {
	Convey("no transactions found", t, func() {
		LogE5Error("", errors.New("error getting transactions"), models.PayableResource{}, validators.PaymentInformation{})
	})
}
