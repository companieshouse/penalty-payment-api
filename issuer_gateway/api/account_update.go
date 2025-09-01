package api

import (
	"fmt"
	"strconv"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api-core/validators"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
)

var getCompanyCodeFromTransaction = utils.GetCompanyCodeFromTransaction

// UpdateIssuerAccountWithPenaltyPaid will update the transactions in E5 as paid.
// resource - is the payable resource from the db representing the penalty(ies)
// payment - is the information about the payment session
func UpdateIssuerAccountWithPenaltyPaid(payableResourceService *services.PayableResourceService,
	client e5.ClientInterface, resource models.PayableResource, payment validators.PaymentInformation, context string) error {
	log.DebugC(context, "converting payment amount from string to float", log.Data{"amount": payment.Amount})
	amountPaid, err := strconv.ParseFloat(payment.Amount, 32)
	if err != nil {
		log.ErrorC(context, err, log.Data{"payment_reference": payment.Reference, "amount": payment.Amount})
		return err
	}

	var transactions []*e5.CreatePaymentTransaction

	for _, t := range resource.Transactions {
		transactions = append(transactions, &e5.CreatePaymentTransaction{
			TransactionReference: t.PenaltyRef,
			Value:                t.Amount,
		})
	}

	// this will be used for the PUON value in E5. it is referred to as paymentId in their spec. X is prefixed to it
	// so that it doesn't clash with other PUON's from different sources when finance produce their reports - namely
	// ones that begin with 'LP' which signify penalties that have been paid outside the digital service.
	paymentID := "X" + payment.PaymentID

	log.DebugC(context, "getting company code from transaction", log.Data{"transaction": transactions[0]})
	companyCode, err := getCompanyCodeFromTransaction(resource.Transactions)

	if err != nil {
		log.ErrorC(context, fmt.Errorf("error getting company code from transaction: %v", err))
		return err
	}

	// three http requests are needed to mark a transactions as paid. The process is 1) create the payment, 2) authorise
	// the payments and finally 3) confirm the payment. if anyone of these fails, the company account will be locked in
	// E5. Finance have confirmed that it is better to keep these locked as a cleanup process will happen naturally in
	// the working day.
	logData := log.Data{
		"company_code":  companyCode,
		"customer_code": resource.CustomerCode,
		"penalty_ref":   transactions[0].TransactionReference,
		"payable_ref":   resource.PayableRef,
		"payment_id":    payment.PaymentID,
		"e5_puon":       paymentID,
		"total_value":   amountPaid,
	}
	log.DebugC(context, "creating payment in E5", logData)
	err = client.CreatePayment(&e5.CreatePaymentInput{
		CompanyCode:  companyCode,
		CustomerCode: resource.CustomerCode,
		PaymentID:    paymentID,
		TotalValue:   amountPaid,
		Transactions: transactions,
	}, "")

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.CreateAction, context); svcErr != nil {
			log.ErrorC(context, svcErr, log.Data{"payment_id": payment.PaymentID, "payable_ref": resource.PayableRef})
			return err
		}
		private.LogE5Error("failed to create payment in E5", err, resource, payment, context)
		return err
	}

	log.DebugC(context, "authorising payment in E5", logData)
	err = client.AuthorisePayment(&e5.AuthorisePaymentInput{
		CompanyCode:   companyCode,
		PaymentID:     paymentID,
		CardReference: payment.ExternalPaymentID,
		CardType:      payment.CardType,
		Email:         payment.CreatedBy,
	}, "")

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.AuthoriseAction, context); svcErr != nil {
			log.ErrorC(context, svcErr, log.Data{"payment_id": payment.PaymentID, "payable_ref": resource.PayableRef})
			return err
		}
		private.LogE5Error("failed to authorise payment in E5", err, resource, payment, context)
		return err
	}

	log.DebugC(context, "confirming payment in E5", logData)
	err = client.ConfirmPayment(&e5.PaymentActionInput{
		CompanyCode: companyCode,
		PaymentID:   paymentID,
	}, context)

	if err != nil {
		if svcErr := RecordIssuerCommandError(payableResourceService, resource, e5.ConfirmAction, context); svcErr != nil {
			log.ErrorC(context, svcErr, log.Data{"payment_id": payment.PaymentID, "payable_ref": resource.PayableRef})
			return err
		}
		private.LogE5Error("failed to confirm payment in E5", err, resource, payment, context)
		return err
	}

	log.InfoC(context, "marked penalty transaction(s) as paid in E5", logData)

	return nil
}

// RecordIssuerCommandError will mark the resource as having failed to update E5.
func RecordIssuerCommandError(payableResourceService *services.PayableResourceService,
	resource models.PayableResource, action e5.Action, context string) error {
	return payableResourceService.DAO.SaveE5Error(resource.CustomerCode, resource.PayableRef, context, action)
}
