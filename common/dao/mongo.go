package dao

import (
	"context"
	"errors"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
)

var client *mongo.Client

func getMongoClient(mongoDBURL string) *mongo.Client {
	if client != nil {
		return client
	}

	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(mongoDBURL)
	client, err := mongo.Connect(ctx, clientOptions)

	// assume the caller of this func cannot handle the case where there is no database connection so the prog must
	// crash here as the service cannot continue.
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// check we can connect to the mongodb mongoInstance. failure here should result in a crash.
	pingContext, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	err = client.Ping(pingContext, nil)
	if err != nil {
		log.Error(errors.New("ping to mongodb timed out. please check the connection to mongodb and that it is running"))
		os.Exit(1)
	}

	log.Info("connected to mongodb successfully")

	return client
}

// MongoDatabaseInterface is an interface that describes the mongodb driver
type MongoDatabaseInterface interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

func getMongoDatabase(mongoDBURL, databaseName string) MongoDatabaseInterface {
	return getMongoClient(mongoDBURL).Database(databaseName)
}

// MongoPayableResourceService is an implementation of the PayableResourceDaoService interface using
// MongoDB as the backend driver.
type MongoPayableResourceService struct {
	db             MongoDatabaseInterface
	CollectionName string
}

// MongoAccountPenaltiesService is an implementation of the AccountPenaltiesDaoService interface using
// MongoDB as the backend driver.
type MongoAccountPenaltiesService struct {
	db             MongoDatabaseInterface
	CollectionName string
}

// CreateAccountPenalties creates a new document in the account_penalties database collection if a
// document does not already exist for the customer
func (m *MongoAccountPenaltiesService) CreateAccountPenalties(dao *models.AccountPenaltiesDao) error {
	log.Info("creating new document in account_penalties collection", log.Data{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	})

	filter := bson.M{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
			"created_at":    dao.CreatedAt,
			"closed_at":     dao.ClosedAt,
			"data":          dao.AccountPenalties,
		},
	}

	opts := options.Update().SetUpsert(true)

	collection := m.db.Collection(m.CollectionName)

	// this allows the creation of the new entry if the document does not already exist to be
	// completed in an atomic operation
	result, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Error(err, log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
		return err
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 1 {
		log.Info("created new document in account_penalties collection", log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
	} else {
		log.Info("no new document created in account_penalties collection as one already exists", log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
	}

	return nil
}

// GetAccountPenalties gets the account penalties from the account_penalties database collection
func (m *MongoAccountPenaltiesService) GetAccountPenalties(customerCode string, companyCode string) (*models.AccountPenaltiesDao, error) {
	log.Info("retrieving document in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
	})

	var resource models.AccountPenaltiesDao

	collection := m.db.Collection(m.CollectionName)
	dbResource := collection.FindOne(context.Background(), bson.M{
		"customer_code": customerCode,
		"company_code":  companyCode,
	})

	err := dbResource.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Debug("no document found in account_penalties collection", log.Data{
				"customer_code": customerCode,
				"company_code":  companyCode,
			})
			return nil, nil
		}
		log.Error(err, log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
		})
		return nil, err
	}

	err = dbResource.Decode(&resource)

	if err != nil {
		log.Error(err, log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
		})
		return nil, err
	}

	return &resource, nil
}

// UpdateAccountPenaltyAsPaid will update the penalty status of an item in account_penalties database collection
func (m *MongoAccountPenaltiesService) UpdateAccountPenaltyAsPaid(customerCode string, companyCode string, penaltyRef string) error {
	log.Info("updating penalty as paid in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
		"penalty_ref":   penaltyRef,
	})

	filter := bson.M{
		"customer_code":              customerCode,
		"company_code":               companyCode,
		"data.transaction_reference": penaltyRef,
	}

	closedAt := time.Now().Truncate(time.Millisecond)

	update := bson.D{
		{
			"$set", bson.M{"data.$.is_paid": true, "closed_at": closedAt},
		},
	}

	collection := m.db.Collection(m.CollectionName)

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err, log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
			"penalty_ref":   penaltyRef,
		})
		return err
	}

	log.Info("successfully updated penalty as paid in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
		"closed_at":     closedAt,
		"penalty_ref":   penaltyRef,
	})

	return nil
}

// DeleteAccountPenalties deletes an entry from the account_penalties database collection
func (m *MongoAccountPenaltiesService) DeleteAccountPenalties(customerCode string, companyCode string) error {
	log.Info("deleting document in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
	})

	filter := bson.M{"customer_code": customerCode, "company_code": companyCode}

	collection := m.db.Collection(m.CollectionName)

	_, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		log.Error(err, log.Data{"customer_code": customerCode, "company_code": companyCode})
		return err
	}

	log.Info("successfully deleted document in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
	})

	return nil
}

// SaveE5Error will update the resource by flagging an error in e5 for a particular action
func (m *MongoPayableResourceService) SaveE5Error(customerCode, payableRef string, action e5.Action) error {
	dao, err := m.GetPayableResource(customerCode, payableRef)
	if err != nil {
		log.Error(err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
		return err
	}

	filter := bson.M{"_id": dao.ID}
	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "e5_command_error", Value: string(action)},
			},
		},
	}

	collection := m.db.Collection(m.CollectionName)

	log.Debug("updating e5 command error in mongo document", log.Data{"_id": dao.ID})

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err, log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})
		return err
	}

	return nil
}

// CreatePayableResource will store the payable request into the database
func (m *MongoPayableResourceService) CreatePayableResource(dao *models.PayableResourceDao) error {

	dao.ID = primitive.NewObjectID()

	collection := m.db.Collection(m.CollectionName)
	_, err := collection.InsertOne(context.Background(), dao)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// GetPayableResource gets the payable request from the database
func (m *MongoPayableResourceService) GetPayableResource(customerCode, payableRef string) (*models.PayableResourceDao, error) {
	var resource models.PayableResourceDao

	collection := m.db.Collection(m.CollectionName)
	dbResource := collection.FindOne(context.Background(), bson.M{"payable_ref": payableRef, "customer_code": customerCode})

	err := dbResource.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Debug("no payable resource found", log.Data{"customer_code": customerCode, "payable_ref": payableRef})
			return nil, nil
		}
		log.Error(err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
		return nil, err
	}

	err = dbResource.Decode(&resource)

	if err != nil {
		log.Error(err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
		return nil, err
	}

	return &resource, nil
}

// UpdatePaymentDetails will save the document back to Mongo
func (m *MongoPayableResourceService) UpdatePaymentDetails(dao *models.PayableResourceDao) error {
	filter := bson.M{"_id": dao.ID}

	update := bson.D{
		{
			Key: "$set", Value: bson.D{
				{Key: "data.payment.status", Value: dao.Data.Payment.Status},
				{Key: "data.payment.reference", Value: dao.Data.Payment.Reference},
				{Key: "data.payment.paid_at", Value: dao.Data.Payment.PaidAt},
				{Key: "data.payment.amount", Value: dao.Data.Payment.Amount},
			},
		},
	}

	collection := m.db.Collection(m.CollectionName)

	log.Debug("updating payment details in mongo document", log.Data{"_id": dao.ID})

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err, log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})
		return err
	}

	log.Debug("updated payment details in mongo document", log.Data{"_id": dao.ID})

	return nil
}

// Shutdown is a hook that can be used to clean up db resources
func (m *MongoPayableResourceService) Shutdown() {
	if client != nil {
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("disconnected from mongodb successfully")
	}
}
