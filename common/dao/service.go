package dao

import (
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
)

// PayableResourceDaoService interface declares how to interact with the persistence layer regardless of underlying technology
type PayableResourceDaoService interface {
	// CreatePayableResource will persist a newly created resource
	CreatePayableResource(dao *models.PayableResourceDao, requestId string) error
	// GetPayableResource will find a single payable resource with the given customerCode and payableRef
	GetPayableResource(customerCode, payableRef string, requestId string) (*models.PayableResourceDao, error)
	// UpdatePaymentDetails will update the resource with changed values
	UpdatePaymentDetails(dao *models.PayableResourceDao, requestId string) error
	// SaveE5Error stored which command to E5 failed e.g. create, authorise or confirm
	SaveE5Error(customerCode, payableRef string, requestId string, action e5.Action) error
	// Shutdown can be called to clean up any open resources that the service may be holding on to.
	Shutdown()
}

var getMongoDB = getMongoDatabase

// NewPayableResourcesDaoService will create a new instance of the PayableResourceDaoService interface. All details about its implementation and the
// database driver will be hidden from outside of this package
func NewPayableResourcesDaoService(cfg *config.Config) PayableResourceDaoService {
	database := getMongoDB(cfg.MongoDBURL, cfg.Database)
	return &MongoPayableResourceService{
		db:             database,
		CollectionName: cfg.PayableResourcesCollection,
	}
}

// AccountPenaltiesDaoService interface declares how to interact with the persistence layer
// regardless of underlying technology
type AccountPenaltiesDaoService interface {
	// CreateAccountPenalties will persist a newly created resource
	CreateAccountPenalties(dao *models.AccountPenaltiesDao, requestId string) error
	// GetAccountPenalties will find the account penalties for a given customerCode and companyCode
	GetAccountPenalties(customerCode string, companyCode string, requestId string) (*models.AccountPenaltiesDao, error)
	// UpdateAccountPenaltyAsPaid will update a transactions as paid for a given customerCode, companyCode and penaltyRef
	UpdateAccountPenaltyAsPaid(customerCode string, companyCode string, penaltyRef string, requestId string) error
	// UpdateAccountPenalties will update the created_at, closed_at and data fields of an existing document
	UpdateAccountPenalties(dao *models.AccountPenaltiesDao, requestId string) error
}

// NewAccountPenaltiesDaoService will create a new instance of the AccountPenaltiesDaoService interface.
// All details about its implementation and the  database driver will be hidden from outside of this package
func NewAccountPenaltiesDaoService(cfg *config.Config) AccountPenaltiesDaoService {
	database := getMongoDB(cfg.MongoDBURL, cfg.Database)
	return &MongoAccountPenaltiesService{
		db:             database,
		CollectionName: cfg.AccountPenaltiesCollection,
	}
}
