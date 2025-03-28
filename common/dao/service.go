package dao

import (
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
)

// Service interface declares how to interact with the persistence layer regardless of underlying technology
type Service interface {
	// CreatePayableResource will persist a newly created resource
	CreatePayableResource(dao *models.PayableResourceDao) error
	// GetPayableResource will find a single payable resource with the given customerCode and payableRef
	GetPayableResource(customerCode, payableRef string) (*models.PayableResourceDao, error)
	// UpdatePaymentDetails will update the resource with changed values
	UpdatePaymentDetails(dao *models.PayableResourceDao) error
	// SaveE5Error stored which command to E5 failed e.g. create, authorise or confirm
	SaveE5Error(customerCode, payableRef string, action e5.Action) error
	// Shutdown can be called to clean up any open resources that the service may be holding on to.
	Shutdown()
}

// NewDAOService will create a new instance of the Service interface. All details about its implementation and the
// database driver will be hidden from outside of this package
func NewDAOService(cfg *config.Config) Service {
	database := getMongoDatabase(cfg.MongoDBURL, cfg.Database)
	return &MongoService{
		db:             database,
		CollectionName: cfg.MongoCollection,
	}
}

// AccountPenaltiesCacheService interface declares how to interact with the persistence layer
// regardless of underlying technology
type AccountPenaltiesCacheService interface {
	// CreateAccountPenalties will persist a newly created resource
	CreateAccountPenalties(dao *AccountPenaltiesDao) error
	// GetAccountPenalties will find the account penalties for a given customerCode
	GetAccountPenalties(customerCode string) (*AccountPenaltiesDao, error)
	// UpdateAccountPenaltyAsPaid will update a transactions as paid for a given customerCode and penaltyReference
	UpdateAccountPenaltyAsPaid(customerCode string, penaltyReference string) error
	// DeleteAccountPenalties will delete the entry for a given customerCode
	DeleteAccountPenalties(customerCode string) error
}

// AccountPenaltiesDao is the persisted resource for account penalties
type AccountPenaltiesDao struct {
	ID               string                    `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt        *time.Time                `json:"created_at" bson:"created_at"`
	AccountPenalties []AccountPenaltiesDataDao `json:"account_penalties" bson:"account_penalties"`
}

// AccountPenaltiesDataDao is the sub document of account penalties
type AccountPenaltiesDataDao struct {
	CompanyCode          string  `json:"company_code" bson:"company_code"`
	LedgerCode           string  `json:"ledger_code" bson:"ledger_code"`
	CustomerCode         string  `json:"customer_code" bson:"customer_code"`
	TransactionReference string  `json:"transaction_reference" bson:"transaction_reference"`
	TransactionDate      string  `json:"transaction_date" bson:"transaction_date"`
	MadeUpDate           string  `json:"made_up_date" bson:"made_up_date"`
	Amount               float64 `json:"amount" bson:"amount"`
	OutstandingAmount    float64 `json:"outstanding_amount" bson:"outstanding_amount"`
	IsPaid               bool    `json:"is_paid" bson:"is_paid"`
	TransactionType      string  `json:"transaction_type" bson:"transaction_type"`
	TransactionSubType   string  `json:"transaction_sub_type" bson:"transaction_sub_type"`
	TypeDescription      string  `json:"type_description" bson:"type_description"`
	DueDate              string  `json:"due_date" bson:"due_date"`
	AccountStatus        string  `json:"account_status" bson:"account_status"`
	DunningStatus        string  `json:"dunning_status" bson:"dunning_status"`
}

// NewAccountPenaltiesDaoService will create a new instance of the AccountPenaltiesCacheService interface.
// All details about its implementation and the  database driver will be hidden from outside of this package
func NewAccountPenaltiesDaoService(cfg *config.Config) AccountPenaltiesCacheService {
	database := getMongoDatabase(cfg.MongoDBURL, cfg.Database)
	return &MongoAccountService{
		db:             database,
		CollectionName: cfg.MongoE5CacheCollection,
	}
}
