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
func AccountPenalties(penaltyRefType, customerCode, companyCode string,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap,
	apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
	cfg, err := getConfig()
	if err != nil {
		log.Error(fmt.Errorf("error getting config: %v", err))
		return nil, services.Error, nil
	}
	log.Debug(fmt.Sprintf("config data: %+v", cfg))

	companyInfoLogData := log.Data{"customer_code": customerCode, "company_code": companyCode}

	log.Info("getting account penalties from cache", companyInfoLogData)
	accountPenalties, err := apDaoSvc.GetAccountPenalties(customerCode, companyCode)

	if accountPenalties == nil {
		log.Info("account penalties not found in cache, getting account penalties from E5 transactions", companyInfoLogData)
		accountPenalties, err = getAccountPenaltiesFromE5Transactions(customerCode, companyCode, cfg, apDaoSvc, false)
	} else if isStale(accountPenalties, cfg) {
		log.Info("account penalties cache record is stale, getting account penalties from E5 transactions", companyInfoLogData)
		accountPenalties, err = getAccountPenaltiesFromE5Transactions(customerCode, companyCode, cfg, apDaoSvc, true)
	}
	if err != nil {
		return nil, services.Error, err
	}

	// Generate the CH preferred format of the results i.e. classify the transactions into
	// payable "penalty" types or non-payable "other" types
	generatedTransactionListFromAccountPenalties, err :=
		generateTransactionList(accountPenalties, penaltyRefType, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error generating transaction list from account penalties: [%v]", err)
		log.Error(err)
		return nil, services.Error, err
	}

	log.Info("Completed AccountPenalties request and mapped to CH penalty transactions", companyInfoLogData)
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

func getAccountPenaltiesFromE5Transactions(
	customerCode string, companyCode string, cfg *config.Config, apDaoSvc dao.AccountPenaltiesDaoService, cacheRecordExists bool) (*models.AccountPenaltiesDao, error) {
	e5Response, err := getTransactionListFromE5(customerCode, companyCode, cfg)
	logData := log.Data{"customer_code": customerCode, "company_code": companyCode}
	if err != nil {
		log.Error(fmt.Errorf("error getting transaction list: [%v]", err))
		return nil, err
	}
	log.Debug("E5 transactions", log.Data{"transactions": e5Response.Transactions})

	if len(e5Response.Transactions) == 0 {
		log.Info("E5 transactions empty, account penalties not cached", logData)
		// If company or transactions do not exist in E5, return account penalties with empty transaction list
		return &models.AccountPenaltiesDao{
			CustomerCode:     customerCode,
			CompanyCode:      companyCode,
			AccountPenalties: make([]models.AccountPenaltiesDataDao, 0),
		}, nil
	} else if cacheRecordExists {
		log.Info("updating account penalties cache from E5 transactions", logData)
		return updateAccountPenaltiesEntry(customerCode, companyCode, e5Response, apDaoSvc), nil
	} else {
		log.Info("creating account penalties cache from E5 transactions", logData)
		return createAccountPenaltiesEntry(customerCode, companyCode, e5Response, apDaoSvc), nil
	}
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
	// If ClosedAt time is set, start counting ttl from then, otherwise, start from CreatedAt
	// Starting from ClosedAt time will ensure that if a user initiates a penalty payment without completing it at the same time
	// and comes back later to complete the payment, we'll have enough confidence that E5 allocation routine
	// would have run before the cache is considered stale and updated. If the ClosedAt time is not set, then we can
	// safely start counting from CreatedAt time.
	ttlStart := accountPenaltiesDao.ClosedAt
	if ttlStart == nil {
		ttlStart = accountPenaltiesDao.CreatedAt
	}

	ttl := getTimeToLive(cfg)
	cacheRecordAge := time.Since(*accountPenaltiesDao.CreatedAt)

	stale := cacheRecordAge == ttl || cacheRecordAge > ttl

	log.Info("Checking if account penalties record is stale ", log.Data{
		"customer_code": accountPenaltiesDao.CustomerCode,
		"company_code":  accountPenaltiesDao.CompanyCode,
		"ttl":           ttl.String(),
		"created_at":    accountPenaltiesDao.CreatedAt,
		"closed_at":     accountPenaltiesDao.ClosedAt,
		"is_stale":      stale,
	})

	return stale
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
