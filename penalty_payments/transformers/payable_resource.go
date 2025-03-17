package transformers

import (
	"fmt"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/constants"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/config"
)

// PayableResourceRequestToDB will take the input request from the REST call and transform it to a dao ready for
// insertion into the database
func PayableResourceRequestToDB(req *models.PayableRequest) *models.PayableResourceDao {
	transactionsDAO := map[string]models.TransactionDao{}
	for _, tx := range req.Transactions {
		transactionsDAO[tx.TransactionID] = models.TransactionDao{
			Amount:     tx.Amount,
			MadeUpDate: tx.MadeUpDate,
			Type:       tx.Type,
			Reason:     tx.Reason,
		}
	}

	reference := utils.GenerateReferenceNumber()
	etag, err := utils.GenerateEtag()
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
	}
	format := "/company/%s/penalties/late-filing/payable/%s"

	self := fmt.Sprintf(format, req.CompanyNumber, reference)

	paymentLinkFormat := "%s/payment"
	paymentLink := fmt.Sprintf(paymentLinkFormat, self)

	resumeJourneyLinkFormat := "/late-filing-penalty/company/%s/penalty/%s/view-penalties"
	resumeJourneyLink := fmt.Sprintf(resumeJourneyLinkFormat, req.CompanyNumber, req.Transactions[0].TransactionID) // Assumes there is only one transaction

	createdAt := time.Now().Truncate(time.Millisecond)
	dao := &models.PayableResourceDao{
		CompanyNumber: req.CompanyNumber,
		Reference:     reference,
		Data: models.PayableResourceDataDao{
			Etag:         etag,
			Transactions: transactionsDAO,
			Payment: models.PaymentDao{
				Status: constants.Pending.String(),
			},
			CreatedAt: &createdAt,
			CreatedBy: models.CreatedByDao{
				Email:    req.CreatedBy.Email,
				ID:       req.CreatedBy.ID,
				Forename: req.CreatedBy.Forename,
				Surname:  req.CreatedBy.Surname,
			},
			Links: models.PayableResourceLinksDao{
				Self:          self,
				Payment:       paymentLink,
				ResumeJourney: resumeJourneyLink,
			},
		},
	}
	return dao
}

// PayableResourceDaoToCreatedResponse will transform a payable resource dao that has successfully been created into
// a http response entity
func PayableResourceDaoToCreatedResponse(model *models.PayableResourceDao) *models.CreatedPayableResource {
	return &models.CreatedPayableResource{
		ID: model.Reference,
		Links: models.CreatedPayableResourceLinks{
			Self: model.Data.Links.Self,
		},
	}
}

// PayableResourceDBToRequest will take the Dao version of a payable resource and convert to a request version
func PayableResourceDBToRequest(payableDao *models.PayableResourceDao) *models.PayableResource {
	var transactions []models.TransactionItem
	for key, val := range payableDao.Data.Transactions {
		tx := models.TransactionItem{
			TransactionID: key,
			Amount:        val.Amount,
			MadeUpDate:    val.MadeUpDate,
			Type:          val.Type,
			Reason:        val.Reason,
		}
		transactions = append(transactions, tx)
	}

	payable := models.PayableResource{
		CompanyNumber: payableDao.CompanyNumber,
		Reference:     payableDao.Reference,
		Transactions:  transactions,
		Etag:          payableDao.Data.Etag,
		CreatedAt:     payableDao.Data.CreatedAt,
		CreatedBy:     models.CreatedBy(payableDao.Data.CreatedBy),
		Links:         models.PayableResourceLinks(payableDao.Data.Links),
		Payment:       models.Payment(payableDao.Data.Payment),
	}
	return &payable
}

// PayableResourceToPaymentDetails will create a PaymentDetails resource (for integrating into payment service) from a PPS PayableResource
func PayableResourceToPaymentDetails(payable *models.PayableResource,
	penaltyDetails config.PenaltyDetails) *models.PaymentDetails {
	var costs []models.Cost
	for _, tx := range payable.Transactions {
		cost := models.Cost{
			Amount:                  fmt.Sprintf("%g", tx.Amount),
			AvailablePaymentMethods: []string{"credit-card"},
			ClassOfPayment:          []string{penaltyDetails.ClassOfPayment},
			Description:             penaltyDetails.Description,
			DescriptionIdentifier:   penaltyDetails.DescriptionId,
			Kind:                    "cost#cost",
			ResourceKind:            penaltyDetails.ResourceKind,
			ProductType:             penaltyDetails.ProductType,
		}
		costs = append(costs, cost)
	}

	payment := models.PaymentDetails{
		Description: penaltyDetails.Description,
		Etag:        payable.Etag, // use the same Etag as PayableResource its built from - if PayableResource changes PaymentDetails may change too
		Kind:        penaltyDetails.ResourceKind,
		Links: models.PaymentDetailsLinks{
			Self:     payable.Links.Payment, // this is the payment details resource so should use payment link from PayableResource
			Resource: payable.Links.Self,    // PayableResources Self link is the resource this PaymentDetails is paying for
		},
		PaidAt:           payable.Payment.PaidAt,
		PaymentReference: payable.Payment.Reference,
		Status:           payable.Payment.Status,
		CompanyNumber:    payable.CompanyNumber,
		Items:            costs,
	}
	return &payment
}
