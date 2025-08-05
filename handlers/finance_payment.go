package handlers

import (
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
)

// FinancePayment interface declares the processing handler for the consumer
type FinancePayment interface {
	ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing, e5PaymentID string,
		cfg *config.Config, isRetry bool) error
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
func (p PenaltyFinancePayment) ProcessFinancialPenaltyPayment(penaltyPayment models.PenaltyPaymentsProcessing,
	e5PaymentID string, cfg *config.Config, isRetry bool) error {
	logContext := log.Data{
		"customer_code":           penaltyPayment.CustomerCode,
		"company_code":            penaltyPayment.CompanyCode,
		"payable_ref":             penaltyPayment.PayableRef,
		"penalty_payment_message": penaltyPayment,
		"e5_payment_id":           e5PaymentID,
		"is_retry":                isRetry,
	}
	log.Info("Financial penalty payment processing started", logContext)

	if isAfterNextDayMidnight(penaltyPayment.CreatedAt) {
		log.Info("Skipping financial penalty payment processing as current time is after next day midnight of created_at", logContext)
		return nil
	}

	var err error

	err = withRetry(cfg, e5.CreateAction, func() error {
		return createPayment(penaltyPayment, p.E5Client, e5PaymentID)
	})
	if err != nil {
		saveE5Error(penaltyPayment, p.PayableResourceDaoService, err, e5PaymentID, e5.CreateAction)
		return err // put it on the retry topic
	}

	err = withRetry(cfg, e5.AuthoriseAction, func() error {
		return authorisePayment(penaltyPayment, p.E5Client, e5PaymentID)
	})
	if err != nil {
		saveE5Error(penaltyPayment, p.PayableResourceDaoService, err, e5PaymentID, e5.AuthoriseAction)
		return nil // don't put it on the retry topic
	}

	err = withRetry(cfg, e5.ConfirmAction, func() error {
		return confirmPayment(penaltyPayment, p.E5Client, e5PaymentID)
	})
	if err != nil {
		saveE5Error(penaltyPayment, p.PayableResourceDaoService, err, e5PaymentID, e5.ConfirmAction)
		return nil // don't put it on the retry topic
	}

	saveE5Success(logContext, p.PayableResourceDaoService, penaltyPayment.CustomerCode, penaltyPayment.PayableRef)
	log.Info("Financial penalty payment processing successful", logContext)
	return nil
}

func isAfterNextDayMidnight(createdAt string) bool {
	parsed, _ := time.Parse(time.RFC3339, createdAt)
	nextDayMidnight := time.Date(parsed.Year(), parsed.Month(), parsed.Day()+1, 0, 0, 0, 0, parsed.Location())
	return time.Now().After(nextDayMidnight)
}

func withRetry(cfg *config.Config, action e5.Action, fn func() error) error {
	attempts := getMaxRetryAttempts(cfg)
	delay := getDelay(cfg)
	maxDelay := getMaxDelay(cfg)

	return retry.Do(
		fn,
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.MaxDelay(maxDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Info("Penalty payment processing retry attempt failed: " + string(action))
		}),
	)
}

func getMaxRetryAttempts(cfg *config.Config) uint {
	var attemptsStr = cfg.PenaltyPaymentsProcessingMaxRetries
	var attempts = uint(3)
	if s, err := strconv.ParseUint(attemptsStr, 10, 32); err == nil {
		attempts = uint(s)
	}
	return attempts
}

func getDelay(cfg *config.Config) time.Duration {
	var delayStr = cfg.PenaltyPaymentsProcessingRetryDelay
	var delay = 1 * time.Second
	if d, err := strconv.ParseInt(delayStr, 10, 32); err == nil {
		delay = time.Duration(d) * time.Second
	}
	return delay
}

func getMaxDelay(cfg *config.Config) time.Duration {
	var maxDelayStr = cfg.PenaltyPaymentsProcessingRetryMaxDelay
	var maxDelay = 5 * time.Second
	if m, err := strconv.ParseUint(maxDelayStr, 10, 32); err == nil {
		maxDelay = time.Duration(m) * time.Second
	}
	return maxDelay
}

func createPayment(penaltyPayment models.PenaltyPaymentsProcessing, client e5.ClientInterface,
	e5PaymentID string) (err error) {
	var e5Transactions []*e5.CreatePaymentTransaction

	for _, t := range penaltyPayment.TransactionPayments {
		e5Transactions = append(e5Transactions, &e5.CreatePaymentTransaction{
			TransactionReference: t.TransactionReference,
			Value:                t.Value,
		})
	}
	err = client.CreatePayment(&e5.CreatePaymentInput{
		CompanyCode:  penaltyPayment.CompanyCode,
		CustomerCode: penaltyPayment.CustomerCode,
		PaymentID:    e5PaymentID,
		TotalValue:   penaltyPayment.TotalValue,
		Transactions: e5Transactions,
	})
	if err != nil {
		return err
	}
	return nil
}

func authorisePayment(penaltyPayment models.PenaltyPaymentsProcessing, client e5.ClientInterface,
	e5PaymentID string) (err error) {
	err = client.AuthorisePayment(&e5.AuthorisePaymentInput{
		CompanyCode:   penaltyPayment.CompanyCode,
		PaymentID:     e5PaymentID,
		CardReference: penaltyPayment.ExternalPaymentID,
		CardType:      penaltyPayment.CardType,
		Email:         penaltyPayment.Email,
	})
	if err != nil {
		return err
	}
	return nil
}

func confirmPayment(penaltyPayment models.PenaltyPaymentsProcessing, client e5.ClientInterface,
	e5PaymentID string) (err error) {
	err = client.ConfirmPayment(&e5.PaymentActionInput{
		CompanyCode: penaltyPayment.CompanyCode,
		PaymentID:   e5PaymentID,
	})
	if err != nil {
		return err
	}
	return nil
}

func saveE5Error(penaltyPayment models.PenaltyPaymentsProcessing, payableResourceDaoService dao.PayableResourceDaoService,
	e5PaymentError error, e5PaymentID string, e5Action e5.Action) {
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

// Ensure that any previous E5 payment errors are cleared following the last successful attempt to confirm payment
func saveE5Success(logContext log.Data, payableResourceDaoService dao.PayableResourceDaoService, customerCode string, payableRef string) {
	if svcErr := payableResourceDaoService.SaveE5Error(customerCode, payableRef, ""); svcErr != nil {
		log.Error(svcErr, logContext)
	}
}
