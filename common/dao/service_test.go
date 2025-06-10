package dao

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

type MockDatabase struct{}

func (m *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return nil
}

func TestUnitService(t *testing.T) {
	original := getMongoDB
	defer func() { getMongoDB = original }()

	getMongoDB = func(mongoDBURL, databaseName string) MongoDatabaseInterface {
		return &MockDatabase{}
	}

	cfg := &config.Config{
		MongoDBURL:                 "mongodb://localhost:27017",
		Database:                   "test",
		PayableResourcesCollection: "payable_resources",
	}

	Convey("successful confirmation", t, func() {
		pr := NewPayableResourcesDaoService(cfg)
		So(pr, ShouldNotBeNil)
	})

	Convey("successful confirmation", t, func() {
		ap := NewAccountPenaltiesDaoService(cfg)
		So(ap, ShouldNotBeNil)
	})
}
