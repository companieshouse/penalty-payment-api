package private

import (
	"errors"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
)

func LogE5Error(message string, originalError error, resource models.PayableResource, payment validators.PaymentInformation) {
	log.Error(errors.New(message), log.Data{
		"payable_ref": resource.PayableRef,
		"payment_id":  payment.PaymentID,
		"amount":      payment.Amount,
		"error":       originalError,
	})
}
