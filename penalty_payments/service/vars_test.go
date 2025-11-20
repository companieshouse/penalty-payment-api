package service

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
)

var customerCode = "NI123456"
var transactionItem = models.TransactionItem{
	PenaltyRef: "A1234567",
}
var payableResource = models.PayableResource{
	CustomerCode: customerCode,
	Transactions: []models.TransactionItem{transactionItem},
}

func setGetCompanyCodeFromTransactionMock(companyCode string) {
	mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return companyCode, nil
	}
	getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction
}

func setGetPenaltyRefTypeFromTransactionMock(penaltyRefType string) {
	mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
		return penaltyRefType, nil
	}
	getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction
}
