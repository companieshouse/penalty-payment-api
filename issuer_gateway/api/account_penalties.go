package api

import (
	"fmt"
	"github.com/companieshouse/penalty-payment-api/common/e5"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
)

var getTransactions = func(companyNumber string, companyCode string, client *e5.Client) (*e5.GetTransactionsResponse, error) {
	return client.GetTransactions(&e5.GetTransactionsInput{CompanyNumber: companyNumber, CompanyCode: companyCode})
}
var getConfig = config.Get
var generateTransactionList = private.GenerateTransactionListFromE5Response

// AccountPenalties is a function that:
// 1. makes a request to e5 to get a list of transactions for the specified company
// 2. takes the results of this request and maps them to a format that the penalty-payment-web can consume
func AccountPenalties(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, services.Error, nil
	}
	client := e5.NewClient(cfg.E5Username, cfg.E5APIURL)
	e5Response, err := getTransactions(companyNumber, companyCode, client)

	if err != nil {
		log.Error(fmt.Errorf("error getting transaction list: [%v]", err))
		return nil, services.Error, err
	}

	// Generate the CH preferred format of the results i.e. classify the transactions into payable "penalty" types or
	// non-payable "other" types
	generatedTransactionListFromE5Response, err :=
		generateTransactionList(e5Response, companyCode, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error generating transaction list from the e5 response: [%v]", err)
		log.Error(err)
		return nil, services.Error, err
	}

	log.Info("Completed AccountPenalties request and mapped to CH penalty transactions", log.Data{"company_number": companyNumber})
	return generatedTransactionListFromE5Response, services.Success, nil
}
