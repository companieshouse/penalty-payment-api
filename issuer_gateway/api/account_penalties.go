package api

import (
	"fmt"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
)

var getTransactions = func(customerCode string, companyCode string, client *e5.Client) (*e5.GetTransactionsResponse, error) {
	return client.GetTransactions(&e5.GetTransactionsInput{CustomerCode: customerCode, CompanyCode: companyCode})
}
var getConfig = config.Get
var generateTransactionList = private.GenerateTransactionListFromAccountPenalties

// AccountPenalties is a function that:
// 1. makes a request to account_penalties collection to get a list of cached transactions for the specified customer
// 2. if no cache entry is found or if the cache entry is stale it makes a request to e5 to get a list of transactions for the specified customer
// 2. takes the results of this request and maps them to a format that the penalty-payment-web can consume
func AccountPenalties(customerCode string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap,
	apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
	accountPenalties, err := apDaoSvc.GetAccountPenalties(customerCode, companyCode)

	cfg, err := getConfig()
	if err != nil {
		log.Error(fmt.Errorf("error getting config: %v", err))
		return nil, services.Error, nil
	}

	if accountPenalties == nil || isStale(accountPenalties, cfg) {
		e5Response, err := getTransactionListFromE5(customerCode, companyCode, cfg)
		if err != nil {
			log.Error(fmt.Errorf("error getting transaction list: [%v]", err))
			return nil, services.Error, err
		}

		if accountPenalties == nil {
			accountPenalties = createAccountPenaltiesEntry(customerCode, companyCode, e5Response, apDaoSvc)
		} else if paymentUpdatedInE5(e5Response, accountPenalties) {
			accountPenalties = updateAccountPenaltiesEntry(customerCode, companyCode, e5Response, apDaoSvc)
		}
	}

	// Generate the CH preferred format of the results i.e. classify the transactions into
	// payable "penalty" types or non-payable "other" types
	generatedTransactionListFromAccountPenalties, err :=
		generateTransactionList(accountPenalties, companyCode, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error generating transaction list from account penalties: [%v]", err)
		log.Error(err)
		return nil, services.Error, err
	}

	log.Info("Completed AccountPenalties request and mapped to CH penalty transactions",
		log.Data{"customer_code": customerCode, "company_code": companyCode})
	return generatedTransactionListFromAccountPenalties, services.Success, nil
}

func createAccountPenaltiesEntry(customerCode string, companyCode string, e5Response *e5.GetTransactionsResponse, apDaoSvc dao.AccountPenaltiesDaoService) *models.AccountPenaltiesDao {
	accountPenalties := convertE5Response(customerCode, companyCode, e5Response)
	err := apDaoSvc.CreateAccountPenalties(&accountPenalties)
	if err != nil {
		log.Error(fmt.Errorf("error creating account penalties: [%v]", err),
			log.Data{"customer_code": customerCode, "company_code": companyCode})
	}

	return &accountPenalties
}

func updateAccountPenaltiesEntry(customerCode string, companyCode string, e5Response *e5.GetTransactionsResponse, apDaoSvc dao.AccountPenaltiesDaoService) *models.AccountPenaltiesDao {
	accountPenalties := convertE5Response(customerCode, companyCode, e5Response)
	err := apDaoSvc.UpdateAccountPenalties(&accountPenalties)
	if err != nil {
		log.Error(fmt.Errorf("error updating account penalties: [%v]", err),
			log.Data{"customer_code": customerCode, "company_code": companyCode})
	}

	return &accountPenalties
}

func getTransactionListFromE5(customerCode string, companyCode string, cfg *config.Config) (*e5.GetTransactionsResponse, error) {
	client := e5.NewClient(cfg.E5Username, cfg.E5APIURL)
	e5Response, err := getTransactions(customerCode, companyCode, client)
	return e5Response, err
}

func convertE5Response(customerCode, companyCode string, response *e5.GetTransactionsResponse) models.AccountPenaltiesDao {
	data := make([]models.AccountPenaltiesDataDao, len(response.Transactions))
	for i, item := range response.Transactions {
		data[i] = models.AccountPenaltiesDataDao{
			CompanyCode:          companyCode,
			LedgerCode:           item.LedgerCode,
			CustomerCode:         customerCode,
			TransactionReference: item.TransactionReference,
			TransactionDate:      item.TransactionDate,
			MadeUpDate:           item.MadeUpDate,
			Amount:               item.Amount,
			OutstandingAmount:    item.OutstandingAmount,
			IsPaid:               item.IsPaid,
			TransactionType:      item.TransactionType,
			TransactionSubType:   item.TransactionSubType,
			TypeDescription:      item.TypeDescription,
			DueDate:              item.DueDate,
			AccountStatus:        item.AccountStatus,
			DunningStatus:        item.DunningStatus,
		}
	}

	createdAt := time.Now().Truncate(time.Millisecond)

	return models.AccountPenaltiesDao{
		CustomerCode:     customerCode,
		CompanyCode:      companyCode,
		CreatedAt:        &createdAt,
		AccountPenalties: data,
	}
}

func isStale(accountPenaltiesDao *models.AccountPenaltiesDao, cfg *config.Config) bool {
	if accountPenaltiesDao.ClosedAt == nil {
		// Penalty is not yet marked as paid in cache so logic to determine if cache record is stale
		// is based on a time to live of 24 hours.

		ttl := getTimeToLive(cfg)
		cacheRecordAge := time.Since(*accountPenaltiesDao.CreatedAt)

		stale := cacheRecordAge == ttl || cacheRecordAge > ttl

		log.Info("Checking if account penalties record is stale ", log.Data{
			"customer_code":       accountPenaltiesDao.CustomerCode,
			"company_code":        accountPenaltiesDao.CompanyCode,
			"TTL":                 ttl.String(),
			"cache_created_since": cacheRecordAge.String(),
			"is_stale":            stale,
		})

		return stale
	} else {
		// Penalty is marked as paid in cache, so logic to determine if cache record is stale is based on
		// E5 allocation routine run.

		e5AllocationRoutineDuration := getE5AllocationRoutineDuration(cfg)
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)                            // 24 hours ago from current time
		e5AllocationRoutineStartHour := cfg.E5AllocationRoutineStartHour // defaults to 0 hour (00:00) if not set in config
		expectedE5AllocationRoutineStartTime := time.Date(
			now.Year(), yesterday.Month(), yesterday.Day(), e5AllocationRoutineStartHour, 0, 0, 0, time.Local)
		expectedE5AllocationRoutineEndTime := expectedE5AllocationRoutineStartTime.Add(e5AllocationRoutineDuration)
		stale := accountPenaltiesDao.ClosedAt.Before(expectedE5AllocationRoutineStartTime) && now.After(expectedE5AllocationRoutineEndTime)

		log.Info("Checking if account penalties record is stale ", log.Data{
			"customer_code":                    accountPenaltiesDao.CustomerCode,
			"company_code":                     accountPenaltiesDao.CompanyCode,
			"e5_allocation_routine_duration":   e5AllocationRoutineDuration.String(),
			"current_time":                     now.Format(time.RFC3339),
			"e5_allocation_routine_start_hour": e5AllocationRoutineStartHour,
			"e5_allocation_routine_start_time": expectedE5AllocationRoutineStartTime.Format(time.RFC3339),
			"e5_allocation_routine_end_time":   expectedE5AllocationRoutineEndTime.Format(time.RFC3339),
			"stale":                            stale,
		})

		// Cache record is considered stale if penalty was marked as paid (and 'ClosedAt' time is) before the start of E5 allocation routine
		// and cache record is assessed after E5 allocation routine has ended
		return stale
	}
}

// This checks that penalties marked as paid in cache are also marked as paid in e5
// Returns false if a penalty marked as paid in cache is not marked as paid in e5, otherwise, returns true
func paymentUpdatedInE5(e5Response *e5.GetTransactionsResponse, accountPenaltiesDao *models.AccountPenaltiesDao) bool {
	var e5Transactions = make(map[string]e5.Transaction)
	for _, transaction := range e5Response.Transactions {
		e5Transactions[transaction.TransactionReference] = transaction
	}

	for _, accountPenalty := range accountPenaltiesDao.AccountPenalties {
		e5Transaction := e5Transactions[accountPenalty.TransactionReference]
		if accountPenalty.IsPaid == true && e5Transaction.IsPaid == false {
			log.Info("cache will not be updated because penalty is marked as paid in cache but not in E5", log.Data{
				"customer_code":         e5Transaction.CustomerCode,
				"company_code":          e5Transaction.CompanyCode,
				"transaction_reference": e5Transaction.TransactionReference,
			})
			return false
		}
	}

	return true
}

func getTimeToLive(cfg *config.Config) time.Duration {
	ttlString := cfg.AccountPenaltiesTTL
	if ttlString == "" {
		ttlString = "24h" // time to live defaults to 24 hours if not set in config
	}

	ttl, err := time.ParseDuration(ttlString)
	if err != nil {
		log.Error(fmt.Errorf("error parsing account penalties TTL: %v", err))
		log.Info("Applying a TTL of 24 hours")
		ttl = 24 * time.Hour // default to TTL of 24 hours if parsing the config TTL string fails
	}

	return ttl
}

func getE5AllocationRoutineDuration(cfg *config.Config) time.Duration {
	e5AllocationRoutineDurationString := cfg.E5AllocationRoutineDuration
	if e5AllocationRoutineDurationString == "" {
		e5AllocationRoutineDurationString = "4h" // E5 allocation routine duration defaults to 4 hours if not set in config
	}

	e5AllocationRoutineDuration, err := time.ParseDuration(e5AllocationRoutineDurationString)
	if err != nil {
		log.Error(fmt.Errorf("error parsing E5 allocation routine duration: %v", err))
		log.Info("Applying a default duration of 4 hours")
		e5AllocationRoutineDuration = 4 * time.Hour // default to 4 hours if parsing of the config duration string fails
	}

	return e5AllocationRoutineDuration
}
