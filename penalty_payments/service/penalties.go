package service

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
)

// TransactionType Enum Type
type TransactionType int

// Enumeration containing all possible types when mapping e5 transactions
const (
	Penalty TransactionType = 1 + iota
	Other
)

// String representation of transaction types
var transactionTypes = [...]string{
	"penalty",
	"other",
}

func (transactionType TransactionType) String() string {
	return transactionTypes[transactionType-1]
}

var getTransactions = func(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap, client *e5.Client) (*e5.GetTransactionsResponse, error) {
	return client.GetTransactions(&e5.GetTransactionsInput{CompanyNumber: companyNumber, CompanyCode: companyCode})
}

// GetPenalties is a function that:
// 1. makes a request to e5 to get a list of transactions for the specified company
// 2. takes the results of this request and maps them to a format that the penalty-payment-web can consume
func GetPenalties(companyNumber string, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
	allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, services.ResponseType, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, services.Error, nil
	}
	client := e5.NewClient(cfg.E5Username, cfg.E5APIURL)
	e5Response, err := getTransactions(companyNumber, companyCode, penaltyDetailsMap, client)

	if err != nil {
		log.Error(fmt.Errorf("error getting transaction list: [%v]", err))
		return nil, services.Error, err
	}

	// Generate the CH preferred format of the results i.e. classify the transactions into payable "penalty" types or
	// non-payable "other" types
	generatedTransactionListFromE5Response, err :=
		generateTransactionListFromE5Response(e5Response, companyCode, penaltyDetailsMap, allowedTransactionsMap)
	if err != nil {
		err = fmt.Errorf("error generating transaction list from the e5 response: [%v]", err)
		log.Error(err)
		return nil, services.Error, err
	}

	log.Info("Completed GetPenalties request and mapped to CH penalty transactions", log.Data{"company_number": companyNumber})
	return generatedTransactionListFromE5Response, services.Success, nil
}

func generateTransactionListFromE5Response(e5Response *e5.GetTransactionsResponse, companyCode string,
	penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, error) {
	// Next, map results to a format that can be used by PPS web
	payableTransactionList := models.TransactionListResponse{}
	etag, err := utils.GenerateEtag()
	if err != nil {
		err = fmt.Errorf("error generating etag: [%v]", err)
		log.Error(err)
		return nil, err
	}

	payableTransactionList.Etag = etag
	payableTransactionList.TotalResults = e5Response.Page.TotalElements

	// Loop through e5 response and construct CH resources
	for _, e5Transaction := range e5Response.Transactions {
		listItem := models.TransactionListItem{}
		listItem.ID = e5Transaction.TransactionReference
		listItem.Etag, err = utils.GenerateEtag()
		if err != nil {
			err = fmt.Errorf("error generating etag: [%v]", err)
			log.Error(err)
			return nil, err
		}
		listItem.IsPaid = e5Transaction.IsPaid
		listItem.Kind = penaltyDetailsMap.Details[companyCode].ResourceKind
		listItem.IsDCA = e5Transaction.AccountStatus == "DCA"
		listItem.DueDate = e5Transaction.DueDate
		listItem.MadeUpDate = e5Transaction.MadeUpDate
		listItem.TransactionDate = e5Transaction.TransactionDate
		listItem.OriginalAmount = e5Transaction.Amount
		listItem.Outstanding = e5Transaction.OutstandingAmount
		// Each transaction needs to be checked and identified as a 'penalty' or 'other'. This allows penalty-payment-web to determine
		// which transactions are payable. This is done using a yaml file to map payable transactions
		// Check if the transaction is allowed and set to 'penalty' if it is
		if _, ok := allowedTransactionsMap.Types[e5Transaction.TransactionType][e5Transaction.TransactionSubType]; ok {
			listItem.Type = Penalty.String()
		} else {
			listItem.Type = Other.String()
		}
		payableTransactionList.Items = append(payableTransactionList.Items, listItem)
	}
	return &payableTransactionList, nil
}
