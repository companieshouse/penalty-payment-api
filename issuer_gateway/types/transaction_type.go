package types

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
