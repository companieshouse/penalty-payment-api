package service

import (
	"fmt"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/transformers"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
)

// PaymentDetailsService contains the PayableResourceService for updating the resource with payment details after a successful/failed payment
type PaymentDetailsService struct {
	PayableResourceService *services.PayableResourceService // needed when implementing PATCH endpoint from payment-processed-consumer
}

// GetPaymentDetailsFromPayableResource transforms a PayableResource into its corresponding Payment details resource
func (service *PaymentDetailsService) GetPaymentDetailsFromPayableResource(req *http.Request,
	payable *models.PayableResource, penaltyDetails config.PenaltyDetails) (*models.PaymentDetails, services.ResponseType, error) {
	paymentDetails := transformers.PayableResourceToPaymentDetails(payable, penaltyDetails)

	if len(paymentDetails.Items) == 0 {
		err := fmt.Errorf("no items in payment details transformed from payable resource [%s]", payable.Reference)
		log.ErrorR(req, err, log.Data{"company_number": payable.CompanyNumber, "payable_reference": payable.Reference,
			"payable_transactions": payable.Transactions})
		return nil, services.InvalidData, err
	}
	return paymentDetails, services.Success, nil
}
