package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"

	"github.com/companieshouse/penalty-payment-api/transformers"
)

// PaymentDetailsService contains the PayableResourceService for updating the resource with payment details after a successful/failed payment
type PaymentDetailsService struct {
	PayableResourceService *services.PayableResourceService // needed when implementing PATCH endpoint from payment-processed-consumer
}

var paymentInstance *PaymentDetailsService
var oncePaymentInstance sync.Once

func GetPaymentDetailsInstance() *PaymentDetailsService {
	oncePaymentInstance.Do(func() {
		paymentInstance = &PaymentDetailsService{
			PayableResourceService: services.GetPayableResourceInstance(),
		}
	})
	return paymentInstance
}

// GetPaymentDetailsFromPayableResource transforms a PayableResource into its corresponding Payment details resource
func (service *PaymentDetailsService) GetPaymentDetailsFromPayableResource(req *http.Request,
	payable *models.PayableResource, penaltyDetailsMap *config.PenaltyDetailsMap, companyCode string) (*models.PaymentDetails, services.ResponseType, error) {
	paymentDetails := transformers.PayableResourceToPaymentDetails(payable, penaltyDetailsMap, companyCode)

	if len(paymentDetails.Items) == 0 {
		err := fmt.Errorf("no items in payment details transformed from payable resource [%s]", payable.Reference)
		log.ErrorR(req, err, log.Data{"company_number": payable.CompanyNumber, "reference": payable.Reference, "payable_transactions": payable.Transactions})
		return nil, services.InvalidData, err
	}
	return paymentDetails, services.Success, nil
}
