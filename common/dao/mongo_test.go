package dao_test

import (
	"context"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestSaveE5Error(t *testing.T) {
	mockDB := new(MockMongoDatabase)
	mockCollection := new(MockMongoCollection)
	mockE5Service := new(mocks.MockE5Service)

	mockDB.On("Collection", "test_collection").Return(mockCollection)

	service := &MongoPayableResourceService{
		db:             mockDB,
		CollectionName: "test_collection",
	}

	dao := &models.PayableResourceDao{
		ID:           primitive.NewObjectID(),
		CompanyNumber: "12345678",
		Reference:     "REF123",
	}

	mockCollection.On("FindOne", mock.Anything, bson.M{"payable_ref": "REF123", "customer_code": "12345678"}).Return(new(mongo.SingleResult))
	mockCollection.On("UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything).Return(&mongo.UpdateResult{}, nil)

	mockE5Service.On("PerformAction", "TestAction").Return(nil)

	err := service.SaveE5Error("12345678", "REF123", "TestAction")
	assert.NoError(t, err)
	mockCollection.AssertCalled(t, "UpdateOne", mock.Anything, bson.M{"_id": dao.ID}, mock.Anything)
	mockE5Service.AssertCalled(t, "PerformAction", "TestAction")
}