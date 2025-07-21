package handlers

import (
	"time"

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
	payableResourceDaoService := p.PayableResourceDaoService
	e5Client := p.E5Client

	customerCode := penaltyPayment.CustomerCode
	payableRef := penaltyPayment.PayableRef
	payableResource, err := payableResourceDaoService.GetPayableResource(customerCode, payableRef)
	if err != nil {
		return err
	}

	createdAt := payableResource.Data.CreatedAt
	e5CommandError := e5.Action(payableResource.E5CommandError)
	logContext := log.Data{
		"customer_code":           customerCode,
		"company_code":            penaltyPayment.CompanyCode,
		"payable_ref":             payableRef,
		"penalty_payment_message": penaltyPayment,
		"e5_payment_id":           e5PaymentID,
		"created_at":              createdAt,
		"e5_command_error":        e5CommandError,
	}
	if isAfterNextDayMidnight(createdAt) {
		log.Info("Skipping financial penalty payment processing as current time is after next day midnight of created_at", logContext)
		return nil
	}

	log.Info("Financial penalty payment processing started", logContext)

	if e5CommandError == "" || e5.CreateAction == e5CommandError {
		retryCreatePaymentError := retryCreatePayment(penaltyPayment, e5PaymentID, payableResourceDaoService, e5Client)
		if retryCreatePaymentError != nil {
			return retryCreatePaymentError
		}
		e5CommandError = ""
	}

	if e5CommandError == "" || e5.AuthoriseAction == e5CommandError {
		retryAuthorisePaymentError := retryAuthorisePayment(penaltyPayment, e5PaymentID, payableResourceDaoService, e5Client)
		if retryAuthorisePaymentError != nil {
			return retryAuthorisePaymentError
		}
		e5CommandError = ""
	}

	if e5CommandError == "" || e5.ConfirmAction == e5CommandError {
		retryConfirmPaymentError := retryConfirmPayment(penaltyPayment, e5PaymentID, payableResourceDaoService, e5Client)
		if retryConfirmPaymentError != nil {
			return retryConfirmPaymentError
		}
		e5CommandError = ""
		// Need to ensure that any previous E5 payment errors are cleared following the last successful attempt to confirm payment
		logContext["e5_command_error"] = e5CommandError
		saveE5Success(logContext, payableResourceDaoService, customerCode, payableRef)
	}

	log.Info("Financial penalty payment processing successful", logContext)
	return nil
}

func retryCreatePayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface) (createPaymentError error) {
	createPaymentError, createPaymentSuccess := retry(3, time.Second, func() (error, bool) {
		return createPayment(penaltyPayment, payableResourceDaoService, client, e5PaymentID)
	})
	if !createPaymentSuccess {
		return createPaymentError
	}
	return nil
}

func retryAuthorisePayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface) (authorisePaymentError error) {
	authorisePaymentError, authorisePaymentSuccess := retry(3, time.Second, func() (error, bool) {
		return authorisePayment(penaltyPayment, payableResourceDaoService, client, e5PaymentID)
	})
	if !authorisePaymentSuccess {
		return authorisePaymentError
	}
	return nil
}

func retryConfirmPayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string, payableResourceDaoService dao.PayableResourceDaoService, client e5.ClientInterface) (confirmPaymentError error) {
	confirmPaymentError, confirmPaymentSuccess := retry(3, time.Second, func() (error, bool) {
		return confirmPayment(penaltyPayment, payableResourceDaoService, client, e5PaymentID)
	})
	if !confirmPaymentSuccess {
		return confirmPaymentError
	}
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
	return nil, true
}

func saveE5Error(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService, e5PaymentError error, e5PaymentID string, e5Action e5.Action) {
	logContext := log.Data{
		"customer_code": penaltyPayment.CustomerCode,
		"company_code":  penaltyPayment.CompanyCode,
		"payable_ref":   penaltyPayment.PayableRef,
		"e5_payment_id": e5PaymentID,
		"e5_action":     e5Action,
	}
	log.Error(e5PaymentError, logContext)
	if svcErr := payableResourceDaoService.SaveE5Error(penaltyPayment.CustomerCode, penaltyPayment.PayableRef, e5Action); svcErr != nil {
		log.Error(svcErr, logContext)
	}
}

// Need to ensure that any previous E5 payment errors are cleared following the last successful attempt to confirm payment
func saveE5Success(logContext log.Data, payableResourceDaoService dao.PayableResourceDaoService, customerCode string, payableRef string) {
	if svcErr := payableResourceDaoService.SaveE5Error(customerCode, payableRef, ""); svcErr != nil {
		log.Error(svcErr, logContext)
	}
}

func isAfterNextDayMidnight(createdAt *time.Time) bool {
	nextDayMidnight := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day()+1, 0, 0, 0, 0, createdAt.Location())
	return time.Now().After(nextDayMidnight)
}

func retry(attempts int, sleep time.Duration, f func() (error, bool)) (error, bool) {
	var err error
	var success bool
	for i := 0; i < attempts; i++ {
		err, success = f()
		if success {
			return nil, true
		}
		time.Sleep(sleep)
	}
	return err, false
}
