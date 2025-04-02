package dao

import (
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
		CollectionName: cfg.PayableResourcesCollection,
	}
}

// AccountPenaltiesCacheService interface declares how to interact with the persistence layer
// regardless of underlying technology
type AccountPenaltiesCacheService interface {
	// CreateAccountPenalties will persist a newly created resource
	CreateAccountPenalties(dao *models.AccountPenaltiesDao) error
	// GetAccountPenalties will find the account penalties for a given customerCode and companyCode
	GetAccountPenalties(customerCode string, companyCode string) (*models.AccountPenaltiesDao, error)
	// UpdateAccountPenaltyAsPaid will update a transactions as paid for a given customerCode, companyCode and penaltyReference
	UpdateAccountPenaltyAsPaid(customerCode string, companyCode string, penaltyReference string) error
	// DeleteAccountPenalties will delete the entry for a given customerCode and companyCode
	DeleteAccountPenalties(customerCode string, companyCode string) error
}

// NewAccountPenaltiesDaoService will create a new instance of the AccountPenaltiesCacheService interface.
// All details about its implementation and the  database driver will be hidden from outside of this package
func NewAccountPenaltiesDaoService(cfg *config.Config) AccountPenaltiesCacheService {
	database := getMongoDatabase(cfg.MongoDBURL, cfg.Database)
	return &MongoAccountService{
		db:             database,
		CollectionName: cfg.AccountPenaltiesCollection,
	}
}
