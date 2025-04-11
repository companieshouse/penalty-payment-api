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
// 2. if no cache entry is found it makes a request to e5 to get a list of transactions for the specified customer
// 2. takes the results of this request and maps them to a format that the penalty-payment-web can consume
func AccountPenalties(customerCode string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap,
	apDaoSvc dao.AccountPenaltiesDaoService) (*models.TransactionListResponse, services.ResponseType, error) {
	accountPenalties, err := apDaoSvc.GetAccountPenalties(customerCode, companyCode)

	if accountPenalties == nil {
		cfg, err := getConfig()
		if err != nil {
			return nil, services.Error, nil
		}

		e5Response, err := getTransactionListFromE5(customerCode, companyCode, cfg)
		if err != nil {
			log.Error(fmt.Errorf("error getting transaction list: [%v]", err))
			return nil, services.Error, err
		}

		accountPenalties = createAccountPenaltiesEntry(customerCode, companyCode, e5Response, apDaoSvc)
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

func createAccountPenaltiesEntry(customerCode string, companyCode string, e5Response *e5.GetTransactionsResponse,
	apDaoSvc dao.AccountPenaltiesDaoService) *models.AccountPenaltiesDao {
	convertedResponse := convertE5Response(customerCode, companyCode, e5Response)
	err := apDaoSvc.CreateAccountPenalties(&convertedResponse)
	if err != nil {
		log.Error(fmt.Errorf("error creating account penalties: [%v]", err),
			log.Data{"customer_code": customerCode, "company_code": companyCode})
	}
	return &convertedResponse
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
