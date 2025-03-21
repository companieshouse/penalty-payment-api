package api

import (
	"strconv"

	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/utils"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
)

var getCompanyCodeFromTransaction = func(transactions []models.TransactionItem) (string, error) {
	return utils.GetCompanyCodeFromTransaction(transactions)
}

// UpdateIssuerAccountWithPenaltyPaid will update the transactions in E5 as paid.
// resource - is the payable resource from the db representing the penalty(ies)
// payment - is the information about the payment session
func UpdateIssuerAccountWithPenaltyPaid(payableResourceService *services.PayableResourceService,
	client *e5.Client, resource models.PayableResource, payment validators.PaymentInformation) error {
	amountPaid, err := strconv.ParseFloat(payment.Amount, 32)
	if err != nil {
		log.Error(err, log.Data{"payment_reference": payment.Reference, "amount": payment.Amount})
		return err
	}

	var transactions []*e5.CreatePaymentTransaction

	for _, t := range resource.Transactions {
		transactions = append(transactions, &e5.CreatePaymentTransaction{
			Reference: t.TransactionID,
			Value:     t.Amount,
		})
	}

	// this will be used for the PUON value in E5. it is referred to as paymentId in their spec. X is prefixed to it
	// so that it doesn't clash with other PUON's from different sources when finance produce their reports - namely
	// ones that begin with 'LP' which signify penalties that have been paid outside of the digital service.
	paymentID := "X" + payment.PaymentID

	companyCode, err := getCompanyCodeFromTransaction(resource.Transactions)

	if err != nil {
		log.Error(err)
		return err
	}

	// three http requests are needed to mark a transactions as paid. The process is 1) create the payment, 2) authorise
	// the payments and finally 3) confirm the payment. if anyone of these fails, the company account will be locked in
	// E5. Finance have confirmed that it is better to keep these locked as a cleanup process will happen naturally in
	// the working day.
	err = client.CreatePayment(&e5.CreatePaymentInput{
		CompanyCode:  companyCode,
		CustomerCode: resource.CustomerCode,
		PaymentID:    paymentID,
		TotalValue:   amountPaid,
		Transactions: transactions,
	})

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.CreateAction); svcErr != nil {
			log.Error(svcErr, log.Data{"payment_id": payment.PaymentID, "payable_reference": resource.Reference})
			return err
		}
		private.LogE5Error("failed to create payment in E5", err, resource, payment)
		return err
	}

	err = client.AuthorisePayment(&e5.AuthorisePaymentInput{
		CompanyCode:   companyCode,
		PaymentID:     paymentID,
		CardReference: payment.ExternalPaymentID,
		CardType:      payment.CardType,
		Email:         payment.CreatedBy,
	})

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.AuthoriseAction); svcErr != nil {
			log.Error(svcErr, log.Data{"payment_id": payment.PaymentID, "payable_reference": resource.Reference})
			return err
		}
		private.LogE5Error("failed to authorise payment in E5", err, resource, payment)
		return err
	}

	err = client.ConfirmPayment(&e5.PaymentActionInput{
		CompanyCode: companyCode,
		PaymentID:   paymentID,
	})

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.ConfirmAction); svcErr != nil {
			log.Error(svcErr, log.Data{"payment_id": payment.PaymentID, "payable_reference": resource.Reference})
			return err
		}
		private.LogE5Error("failed to confirm payment in E5", err, resource, payment)
		return err
	}

	log.Info("marked penalty transaction(s) as paid in E5", log.Data{
		"payable_reference": resource.Reference,
		"payment_id":        payment.PaymentID,
		"e5_puon":           payment.PaymentID,
	})

	return nil
}

// RecordIssuerCommandError will mark the resource as having failed to update E5.
func RecordIssuerCommandError(payableResourceService *services.PayableResourceService,
	resource models.PayableResource, action e5.Action) error {
	return payableResourceService.DAO.SaveE5Error(resource.CustomerCode, resource.Reference, action)
}
