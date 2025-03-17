package services

import (
	"errors"
	"fmt"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/transformers"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/config"
)

var (
	// ErrAlreadyPaid represents when the penalty payable resource is already paid
	ErrAlreadyPaid = errors.New("the Penalty has already been paid")
	// ErrPenaltyNotFound represents when the payable resource does not exist in the db
	ErrPenaltyNotFound = errors.New("the Penalty does not exist")
)

// PayableResourceService contains the DAO for db access
type PayableResourceService struct {
	DAO    dao.Service
	Config *config.Config
}

// GetPayableResource retrieves the payable resource with the given company number and reference from the database
func (s *PayableResourceService) GetPayableResource(req *http.Request, companyNumber string, reference string) (*models.PayableResource, ResponseType, error) {
	payable, err := s.DAO.GetPayableResource(companyNumber, reference)
	if err != nil {
		err = fmt.Errorf("error getting payable resource from db: [%v]", err)
		log.ErrorR(req, err)
		return nil, Error, err
	}
	if payable == nil {
		log.TraceR(req, "payable resource not found", log.Data{"company_number": companyNumber, "payable_reference": reference})
		return nil, NotFound, nil
	}

	payableRest := transformers.PayableResourceDBToRequest(payable)
	return payableRest, Success, nil
}

// UpdateAsPaid will update the resource as paid and persist the changes in the database
func (s *PayableResourceService) UpdateAsPaid(resource models.PayableResource, payment validators.PaymentInformation) error {
	log.Info("update as paid start")
	model, err := s.DAO.GetPayableResource(resource.CompanyNumber, resource.Reference)
	if err != nil {
		err = fmt.Errorf("error getting payable resource from db: [%v]", err)
		log.Error(err, log.Data{
			"payable_reference": resource.Reference,
			"company_number":    resource.CompanyNumber,
		})
		return ErrPenaltyNotFound
	}

	// check if this resource has already been paid
	if model.IsPaid() {
		err = errors.New("this penalty has already been paid")
		log.Error(err, log.Data{
			"payable_ref":   model.PayableRef,
			"customer_code": model.CustomerCode,
			"payment_id":    model.Data.Payment.Reference,
		})
		return ErrAlreadyPaid
	}

	model.Data.Payment.Reference = payment.Reference
	model.Data.Payment.Status = payment.Status
	model.Data.Payment.PaidAt = &payment.CompletedAt
	model.Data.Payment.Amount = payment.Amount

	return s.DAO.UpdatePaymentDetails(model)
}
