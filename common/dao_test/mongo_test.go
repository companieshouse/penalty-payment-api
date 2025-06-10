package dao_test

import (
	"context"
	"testing"

	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockMongoDatabase is a mock implementation of the MongoDatabaseInterface
type MockMongoDatabase struct {
	mock.Mock
}

func (m *MockMongoDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	args := m.Called(name)
	return args.Get(0).(*mongo.Collection)
}

// MockMongoCollection is a mock implementation of a MongoDB collection
type MockMongoCollection struct {
	mock.Mock
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockMongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// MockE5Action is a mock implementation of the e5.Action interface
type MockE5Action struct {
	mock.Mock
}

func (m *MockE5Action) String() string {
	args := m.Called()
	return args.String(0)
}

func TestSaveE5Error(t *testing.T) {
	// Mock dependencies
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockE5Action := new(MockE5Action)

	// Set up the mock database and collection
	mockDB.On("Collection", "test_collection").Return(mockCollection)

	// Create the service
	service := &dao.MongoPayableResourceService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	// Create a mock PayableResourceDao
	mockPayableResource := &models.PayableResourceDao{
		ID:           primitive.NewObjectID(),
		CustomerCode: "12345678",
		PayableRef:   "REF123",
	}

	// Mock the behavior of the collection's FindOne and UpdateOne methods
	mockCollection.On("FindOne", mock.Anything, bson.M{"payable_ref": "REF123", "customer_code": "12345678"}).Return(&mongo.SingleResult{})
	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": mockPayableResource.ID}, mock.Anything).Return(&mongo.UpdateResult{}, nil)

	// Mock the behavior of the e5.Action
	mockE5Action.On("String").Return("TestAction")

	// Call the method under test
	err := service.SaveE5Error("12345678", "REF123", mockE5Action)

	// Assertions
	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": mockPayableResource.ID}, mock.Anything)
	mockE5Action.AssertCalled(t, "String")
}