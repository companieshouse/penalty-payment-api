package dao

import (
	"context"
	"errors"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	//"github.com/companieshouse/penalty-payment-api/e5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockMongoDatabase is a mock implementation of MongoDatabaseInterface
type MockMongoDatabase struct {
	mock.Mock
}

func (m *MockMongoDatabase) Collection(name string, opts ...*mongo.CollectionOptions) *mongo.Collection {
	args := m.Called(name)
	return args.Get(0).(*mongo.Collection)
}

// MockMongoCollection is a mock implementation of MongoDB collection
type MockMongoCollection struct {
	mock.Mock
}

func (m *MockMongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockMongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func TestCreatePayableResource(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockDB.On("Collection", "test_collection").Return(mockCollection)

	service := &MongoService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	dao := &models.PayableResourceDao{
		CompanyNumber: "12345678",
		Reference:     "REF123",
	}

	mockCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil)

	err := service.CreatePayableResource(dao)
	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "InsertOne", mock.Anything, mock.Anything)
}

func TestGetPayableResource(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockDB.On("Collection", "test_collection").Return(mockCollection)

	service := &MongoService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	expectedResource := &models.PayableResourceDao{
		ID:           primitive.NewObjectID(),
		CompanyNumber: "12345678",
		Reference:     "REF123",
	}

	mockSingleResult := new(mongo.SingleResult)
	mockSingleResult.On("Decode", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PayableResourceDao)
		*arg = *expectedResource
	})
	mockCollection.On("FindOne", mock.Anything, bson.M{"reference": "REF123", "company_number": "12345678"}).Return(mockSingleResult)

	resource, err := service.GetPayableResource("12345678", "REF123")
	assert.NoError(t, err)
	assert.Equal(t, expectedResource, resource)
}

func TestUpdatePaymentDetails(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockDB.On("Collection", "test_collection").Return(mockCollection)

	service := &MongoService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	dao := &models.PayableResourceDao{
		ID: primitive.NewObjectID(),
		Data: models.PayableResourceData{
			Payment: models.PaymentDetails{
				Status:    "Paid",
				Reference: "PAY123",
				PaidAt:    "2023-10-01T12:00:00Z",
				Amount:    100.0,
			},
		},
	}

	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything).Return(&mongo.UpdateResult{}, nil)

	err := service.UpdatePaymentDetails(dao)
	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything)
}

func TestSaveE5Error(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockDB.On("Collection", "test_collection").Return(mockCollection)

	service := &MongoService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	dao := &models.PayableResourceDao{
		ID:           primitive.NewObjectID(),
		CompanyNumber: "12345678",
		Reference:     "REF123",
	}

	mockCollection.On("FindOne", mock.Anything, bson.M{"reference": "REF123", "company_number": "12345678"}).Return(new(mongo.SingleResult))
	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything).Return(&mongo.UpdateResult{}, nil)

	// err := service.SaveE5Error("12345678", "REF123", e5.Action("TestAction"))
	// assert.NoError(t, err)
	// mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything)
}