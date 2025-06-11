package dao

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/companieshouse/penalty-payment-api/common"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitService(t *testing.T) {
	getMongoDB = func(mongoDBURL, databaseName string) common.MongoDatabaseInterface {
		return &MongoDatabaseWrapper{
			db: &mongo.Database{},
		}
	}

	Convey("successful creation of new payable resources dao service", t, func() {
		cfg := &config.Config{
			MongoDBURL:                 "mongodb://localhost:27017",
			Database:                   "test",
			PayableResourcesCollection: "payable_resources",
		}

		pr := NewPayableResourcesDaoService(cfg)
		So(pr, ShouldNotBeNil)
	})

	Convey("successful creation of new account penalties dao service", t, func() {
		cfg := &config.Config{
			MongoDBURL:                 "mongodb://localhost:27017",
			Database:                   "test",
			AccountPenaltiesCollection: "account_penalties",
		}

		ap := NewAccountPenaltiesDaoService(cfg)
		So(ap, ShouldNotBeNil)
	})
}
