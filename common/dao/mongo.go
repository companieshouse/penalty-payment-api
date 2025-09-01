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
	"github.com/companieshouse/penalty-payment-api/common/interfaces"
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

func getMongoDatabase(mongoDBURL, databaseName string) interfaces.MongoDatabaseInterface {
	db := getMongoClient(mongoDBURL).Database(databaseName)
	return &MongoDatabaseWrapper{db: db}
}

type MongoCollectionWrapper struct {
	collection *mongo.Collection
}

type MongoDatabaseWrapper struct {
	db *mongo.Database
}

func (m *MongoDatabaseWrapper) Collection(name string, opts ...*options.CollectionOptions) interfaces.MongoCollectionInterface {
	return &MongoCollectionWrapper{collection: m.db.Collection(name, opts...)}
}

func (m *MongoCollectionWrapper) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.collection.InsertOne(ctx, document, opts...)
}

func (m *MongoCollectionWrapper) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return m.collection.FindOne(ctx, filter, opts...)
}

func (m *MongoCollectionWrapper) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.collection.UpdateOne(ctx, filter, update, opts...)
}

func (m *MongoCollectionWrapper) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.collection.DeleteOne(ctx, filter, opts...)
}

// MongoPayableResourceService is an implementation of the PayableResourceDaoService interface using
// MongoDB as the backend driver.
type MongoPayableResourceService struct {
	db             interfaces.MongoDatabaseInterface
	CollectionName string
}

// MongoAccountPenaltiesService is an implementation of the AccountPenaltiesDaoService interface using
// MongoDB as the backend driver.
type MongoAccountPenaltiesService struct {
	db             interfaces.MongoDatabaseInterface
	CollectionName string
}

// CreateAccountPenalties creates a new document in the account_penalties database collection if a
// document does not already exist for the customer
func (m *MongoAccountPenaltiesService) CreateAccountPenalties(dao *models.AccountPenaltiesDao, ctx string) error {
	log.InfoC(ctx, "creating new document in account_penalties collection", log.Data{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	})

	filter := bson.M{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	}

	// setOnInsert is used here with SetUpsert below to ensure that if a document exists then it is not updated
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
		log.ErrorC(ctx, err, log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
		return err
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 1 {
		log.InfoC(ctx, "created new document in account_penalties collection", log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
	} else {
		log.InfoC(ctx, "no new document created in account_penalties collection as one already exists", log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
	}

	return nil
}

// GetAccountPenalties gets the account penalties from the account_penalties database collection
func (m *MongoAccountPenaltiesService) GetAccountPenalties(customerCode string, companyCode, ctx string) (*models.AccountPenaltiesDao, error) {
	logContext := log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
	}

	var resource models.AccountPenaltiesDao

	collection := m.db.Collection(m.CollectionName)
	dbResource := collection.FindOne(context.Background(), bson.M{
		"customer_code": customerCode,
		"company_code":  companyCode,
	})

	err := dbResource.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.DebugC(ctx, "no document found in account_penalties collection", logContext)
			return nil, err
		}
		log.ErrorC(ctx, err, logContext)
		return nil, err
	}

	err = dbResource.Decode(&resource)

	if err != nil {
		log.ErrorC(ctx, err, logContext)
		return nil, err
	}

	log.DebugC(ctx, "GetAccountPenalties", logContext, log.Data{"account_penalties": resource})

	return &resource, nil
}

// UpdateAccountPenaltyAsPaid will update the penalty status of an item in account_penalties database collection
func (m *MongoAccountPenaltiesService) UpdateAccountPenaltyAsPaid(customerCode string, companyCode string, penaltyRef, ctx string) error {
	log.InfoC(ctx, "updating penalty as paid in account_penalties collection", log.Data{
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

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.ErrorC(ctx, err, log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
			"penalty_ref":   penaltyRef,
		})
		return err
	}

	if result.ModifiedCount == 0 {
		err = errors.New("failed to update penalty as paid in account_penalties collection as no penalty was found")
		log.ErrorC(ctx, err, log.Data{
			"customer_code": customerCode,
			"company_code":  companyCode,
			"penalty_ref":   penaltyRef,
		})
		return err
	}

	log.InfoC(ctx, "successfully updated penalty as paid in account_penalties collection", log.Data{
		"customer_code": customerCode,
		"company_code":  companyCode,
		"closed_at":     closedAt,
		"penalty_ref":   penaltyRef,
	})

	return nil
}

// UpdateAccountPenalties updates the created_at, closed_at and data fields of an existing document
func (m *MongoAccountPenaltiesService) UpdateAccountPenalties(dao *models.AccountPenaltiesDao, ctx string) error {
	log.InfoC(ctx, "updating existing document in account_penalties collection", log.Data{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	})

	filter := bson.M{
		"customer_code": dao.CustomerCode,
		"company_code":  dao.CompanyCode,
	}

	update := bson.M{
		"$set": bson.M{
			"created_at": dao.CreatedAt,
			"closed_at":  dao.ClosedAt,
			"data":       dao.AccountPenalties,
		},
	}

	collection := m.db.Collection(m.CollectionName)

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.ErrorC(ctx, err, log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
		return err
	}

	if result.ModifiedCount == 1 {
		log.DebugC(ctx, "updated a document in account_penalties collection", log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
	} else {
		err = errors.New("failed to update document in account_penalties collection")
		log.ErrorC(ctx, err, log.Data{
			"customer_code": dao.CustomerCode,
			"company_code":  dao.CompanyCode,
		})
		return err
	}

	return nil
}

// SaveE5Error will update the resource by flagging an error in e5 for a particular action
func (m *MongoPayableResourceService) SaveE5Error(customerCode, payableRef, ctx string, action e5.Action) error {
	dao, err := m.GetPayableResource(customerCode, payableRef, ctx)
	if err != nil {
		log.ErrorC(ctx, err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
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

	log.DebugC(ctx, "updating e5 command error in mongo document", log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef, "e5_command_error": action})

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.ErrorC(ctx, err, log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})
		return err
	}

	return nil
}

// CreatePayableResource will store the payable request into the database
func (m *MongoPayableResourceService) CreatePayableResource(dao *models.PayableResourceDao, ctx string) error {

	dao.ID = primitive.NewObjectID()

	collection := m.db.Collection(m.CollectionName)
	_, err := collection.InsertOne(context.Background(), dao)
	if err != nil {
		log.ErrorC(ctx, err)
		return err
	}

	return nil
}

// GetPayableResource gets the payable request from the database
func (m *MongoPayableResourceService) GetPayableResource(customerCode, payableRef, ctx string) (*models.PayableResourceDao, error) {
	var resource models.PayableResourceDao

	collection := m.db.Collection(m.CollectionName)
	dbResource := collection.FindOne(context.Background(), bson.M{"payable_ref": payableRef, "customer_code": customerCode})

	err := dbResource.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.DebugC(ctx, "no payable resource found", log.Data{"customer_code": customerCode, "payable_ref": payableRef})
			return nil, err
		}
		log.ErrorC(ctx, err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
		return nil, err
	}

	err = dbResource.Decode(&resource)

	if err != nil {
		log.ErrorC(ctx, err, log.Data{"customer_code": customerCode, "payable_ref": payableRef})
		return nil, err
	}

	return &resource, nil
}

// UpdatePaymentDetails will save the document back to Mongo
func (m *MongoPayableResourceService) UpdatePaymentDetails(dao *models.PayableResourceDao, ctx string) error {
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

	log.DebugC(ctx, "updating payment details in mongo document", log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.ErrorC(ctx, err, log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})
		return err
	}

	log.DebugC(ctx, "updated payment details in mongo document", log.Data{"_id": dao.ID, "customer_code": dao.CustomerCode, "payable_ref": dao.PayableRef})

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
