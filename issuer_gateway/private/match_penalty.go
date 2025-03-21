package private

import (
	"errors"
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var (
	ErrTransactionDoesNotExist   = errors.New("invalid transaction")
	ErrTransactionNotPayable     = errors.New("you cannot pay for this type of transaction")
	ErrTransactionDCA            = errors.New("the transaction is with a debt collecting agency")
	ErrTransactionIsPaid         = errors.New("this transaction is already paid")
	ErrTransactionIsPartPaid     = errors.New("the transaction is already part paid")
	ErrTransactionAmountMismatch = errors.New("you can only pay off the full amount of the transaction")
	ErrMultiplePenalties         = errors.New("the company has more than one outstanding penalty")
)

func MatchPenalty(referenceTransactions []models.TransactionListItem,
	transactionToMatch models.TransactionItem,
	customerCode string) (*models.TransactionItem, error) {

	referenceTransactionsMap := mapTransactions(referenceTransactions)
	transactionInfo := map[string]interface{}{
		"transaction_ref": transactionToMatch.TransactionID,
		"customer_code":   customerCode,
	}

	matched, ok := referenceTransactionsMap[transactionToMatch.TransactionID]
	if !ok {
		log.Info("disallowing paying for a transaction that does not exist in E5", transactionInfo)
		return nil, ErrTransactionDoesNotExist
	}

	valid, err := validate(matched, transactionInfo, transactionToMatch)
	if valid {
		matchedPenalty := models.TransactionItem{
			TransactionID: matched.ID,
			Amount:        matched.Outstanding,
			Type:          matched.Type,
			MadeUpDate:    matched.MadeUpDate,
			IsDCA:         matched.IsDCA,
			IsPaid:        matched.IsPaid,
			Reason:        matched.Reason,
		}
		return &matchedPenalty, nil
	} else {
		return nil, err[0]
	}
}

func validate(
	refTransaction models.TransactionListItem,
	data map[string]interface{},
	transactionToMatch models.TransactionItem) (bool, []error) {

	var errs []error
	valid := true

	if refTransaction.IsPartPaid() {
		log.Info("the penalty that is trying to be paid is already part paid", data)
		valid = false
		errs = append(errs, ErrTransactionIsPartPaid)
	}
	if refTransaction.IsPaid {
		log.Info("disallowing paying for a transaction that is already paid", data)
		valid = false
		errs = append(errs, ErrTransactionIsPaid)
	}
	if refTransaction.Type != types.Penalty.String() {
		log.Info("disallowing paying for a transaction that is not a penalty", data)
		valid = false
		errs = append(errs, ErrTransactionNotPayable)
	}
	if refTransaction.Outstanding != transactionToMatch.Amount {
		data["attempted_amount"] = fmt.Sprintf("%f", transactionToMatch.Amount)
		data["outstanding_amount"] = fmt.Sprintf("%f", refTransaction.Outstanding)
		log.Info("disallowing paying for transaction as attempting to pay off partial balance", data)
		valid = false
		errs = append(errs, ErrTransactionAmountMismatch)
	}
	if refTransaction.IsDCA {
		log.Info("the transaction that is trying to be paid is with a debt collecting agency", data)
		valid = false
		errs = append(errs, ErrTransactionDCA)
	}

	return valid, errs
}

func mapTransactions(transactionListItems []models.TransactionListItem) map[string]models.TransactionListItem {

	itemMap := map[string]models.TransactionListItem{}
	for _, tx := range transactionListItems {
		itemMap[tx.ID] = tx
	}

	return itemMap
}
