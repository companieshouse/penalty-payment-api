package handlers

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
)

// FinancePayment interface declares the processing handler for the consumer
type FinancePayment interface {
	ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string) error
}

// PenaltyFinancePayment is the processing handler for the consumer
type PenaltyFinancePayment struct {
	E5Client                  e5.ClientInterface
	PayableResourceDaoService dao.PayableResourceDaoService
}

// ProcessFinancialPenaltyPayment will update the transactions in E5 as paid.
// Three http requests are needed to mark a transactions as paid. The process is 1) create the payment, 2) authorise
// the payments and finally 3) confirm the payment. If any one of these fails, the company account will be locked in
// E5. Finance have confirmed that it is better to keep these locked as a cleanup process will happen naturally in
// the working day.
func (p PenaltyFinancePayment) ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string) error {
	log.Info("Financial penalty payment processing started", log.Data{
		"e5_payment_id":         e5PaymentID,
		"attempt":               penaltyPayment.Attempt,
		"company_code":          penaltyPayment.CompanyCode,
		"customer_code":         penaltyPayment.CustomerCode,
		"payment_id":            penaltyPayment.PaymentID,
		"external_payment_id":   penaltyPayment.ExternalPaymentID,
		"payment_reference":     penaltyPayment.PaymentReference,
		"payment_amount":        penaltyPayment.PaymentAmount,
		"total_value":           penaltyPayment.TotalValue,
		"transaction_reference": penaltyPayment.TransactionPayments[0].TransactionReference,
		"value":                 penaltyPayment.TransactionPayments[0].Value,
		"card_type":             penaltyPayment.CardType,
		"email":                 penaltyPayment.Email,
		"payable_ref":           penaltyPayment.PayableRef,
	})

	createPaymentError, createPaymentSuccess := createPayment(penaltyPayment, p.PayableResourceDaoService, p.E5Client, e5PaymentID)
	if !createPaymentSuccess {
		return createPaymentError
	}

	authorisePaymentError, authorisePaymentSuccess := authorisePayment(penaltyPayment, p.PayableResourceDaoService, p.E5Client, e5PaymentID)
	if !authorisePaymentSuccess {
		return authorisePaymentError
	}

	confirmPaymentError, confirmPaymentSuccess := confirmPayment(penaltyPayment, p.PayableResourceDaoService, p.E5Client, e5PaymentID)
	if !confirmPaymentSuccess {
		return confirmPaymentError
	}

	log.Info("Financial penalty payment processing successful", log.Data{"e5_payment_id": e5PaymentID, "customer_code": penaltyPayment.CustomerCode, "company_code": penaltyPayment.CompanyCode, "payable_ref": penaltyPayment.PayableRef})
	return nil
}

func createPayment(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface, e5PaymentID string) (createPaymentError error, createPaymentSuccess bool) {
	var e5Transactions []*e5.CreatePaymentTransaction

	for _, t := range penaltyPayment.TransactionPayments {
		e5Transactions = append(e5Transactions, &e5.CreatePaymentTransaction{
			TransactionReference: t.TransactionReference,
			Value:                t.Value,
		})
	}
	createPaymentError = client.CreatePayment(&e5.CreatePaymentInput{
		CompanyCode:  penaltyPayment.CompanyCode,
		CustomerCode: penaltyPayment.CustomerCode,
		PaymentID:    e5PaymentID,
		TotalValue:   penaltyPayment.TotalValue,
		Transactions: e5Transactions,
	})
	if createPaymentError != nil {
		saveE5Error(penaltyPayment, payableResourceDaoService, createPaymentError, e5PaymentID, e5.CreateAction)
		return createPaymentError, false
	}
	return nil, true
}

func authorisePayment(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface, e5PaymentID string) (authorisePaymentError error, authorisePaymentSuccess bool) {
	authorisePaymentError = client.AuthorisePayment(&e5.AuthorisePaymentInput{
		CompanyCode:   penaltyPayment.CompanyCode,
		PaymentID:     e5PaymentID,
		CardReference: penaltyPayment.ExternalPaymentID,
		CardType:      penaltyPayment.CardType,
		Email:         penaltyPayment.Email,
	})
	if authorisePaymentError != nil {
		saveE5Error(penaltyPayment, payableResourceDaoService, authorisePaymentError, e5PaymentID, e5.AuthoriseAction)
		return authorisePaymentError, false
	}
	return nil, true
}

func confirmPayment(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface, e5PaymentID string) (confirmPaymentError error, confirmPaymentSuccess bool) {
	confirmPaymentError = client.ConfirmPayment(&e5.PaymentActionInput{
		CompanyCode: penaltyPayment.CompanyCode,
		PaymentID:   e5PaymentID,
	})
	if confirmPaymentError != nil {
		saveE5Error(penaltyPayment, payableResourceDaoService, confirmPaymentError, e5PaymentID, e5.ConfirmAction)
		return confirmPaymentError, false
	}
	return nil, false
}

func saveE5Error(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService, e5PaymentError error, e5PaymentID string, e5Action e5.Action) {
	logContext := log.Data{"customer_code": penaltyPayment.CustomerCode, "company_code": penaltyPayment.CompanyCode, "payable_ref": penaltyPayment.PayableRef, "e5_payment_id": e5PaymentID, "e5_action": e5Action}
	log.Error(e5PaymentError, logContext)
	if svcErr := payableResourceDaoService.SaveE5Error(penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5Action); svcErr != nil {
		log.Error(svcErr, logContext)
	}
}
