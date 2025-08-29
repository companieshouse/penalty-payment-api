package private

import (
	"errors"
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
)

var (
	ErrPenaltyDoesNotExist   = errors.New("invalid penalty")
	ErrPenaltyNotPayable     = errors.New("you cannot pay for this type of penalty")
	ErrPenaltyDCA            = errors.New("the penalty is with a debt collecting agency")
	ErrPenaltyIsPaid         = errors.New("this penalty is already paid")
	ErrPenaltyIsPartPaid     = errors.New("the penalty is already part paid")
	ErrPenaltyAmountMismatch = errors.New("you can only pay off the full amount of the penalty")
)

func MatchPenalty(referenceTransactions []models.TransactionListItem,
	transactionToMatch models.TransactionItem, customerCode, context string) (*models.TransactionItem, error) {

	referenceTransactionsMap := mapTransactions(referenceTransactions)
	transactionInfo := map[string]interface{}{
		"penalty_ref":   transactionToMatch.PenaltyRef,
		"customer_code": customerCode,
	}

	log.DebugC(context, "checking if penalty is payable", transactionInfo)

	matched, ok := referenceTransactionsMap[transactionToMatch.PenaltyRef]
	if !ok {
		log.InfoC(context, "disallowing paying for a penalty that does not exist in E5", transactionInfo)
		return nil, ErrPenaltyDoesNotExist
	}

	valid, err := validate(matched, transactionInfo, transactionToMatch, context)
	if valid {
		matchedPenalty := models.TransactionItem{
			PenaltyRef: matched.ID,
			Amount:     matched.Outstanding,
			Type:       matched.Type,
			MadeUpDate: matched.MadeUpDate,
			IsDCA:      matched.IsDCA,
			IsPaid:     matched.IsPaid,
			Reason:     matched.Reason,
		}
		log.DebugC(context, "penalty is payable", log.Data{"penalty": matchedPenalty})
		return &matchedPenalty, nil
	} else {
		return nil, err[0]
	}
}

func validate(
	refTransaction models.TransactionListItem,
	data map[string]interface{},
	transactionToMatch models.TransactionItem, context string) (bool, []error) {

	var errs []error
	valid := true

	if refTransaction.IsPartPaid() {
		log.InfoC(context, "attempting to pay a penalty that is already part paid", data)
		valid = false
		errs = append(errs, ErrPenaltyIsPartPaid)
	}
	if refTransaction.IsPaid {
		log.InfoC(context, "disallowing paying for a penalty that is already paid", data)
		valid = false
		errs = append(errs, ErrPenaltyIsPaid)
	}
	if refTransaction.Type != types.Penalty.String() {
		log.InfoC(context, "disallowing paying for a transaction that is not a penalty", data)
		valid = false
		errs = append(errs, ErrPenaltyNotPayable)
	}
	if refTransaction.Outstanding != transactionToMatch.Amount {
		data["attempted_amount"] = fmt.Sprintf("%f", transactionToMatch.Amount)
		data["outstanding_amount"] = fmt.Sprintf("%f", refTransaction.Outstanding)
		log.InfoC(context, "attempting to pay off partial balance of a penalty", data)
		valid = false
		errs = append(errs, ErrPenaltyAmountMismatch)
	}
	if refTransaction.IsDCA {
		log.InfoC(context, "attempting to pay a penalty that is with a debt collecting agency", data)
		valid = false
		errs = append(errs, ErrPenaltyDCA)
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
