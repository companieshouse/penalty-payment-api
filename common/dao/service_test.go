package dao

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/companieshouse/penalty-payment-api/common/interfaces"
	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

var dbUrl = "mongodb://localhost:27017"
var db = "test"

func TestUnitService(t *testing.T) {
	getMongoDB = func(mongoDBURL, databaseName string) interfaces.MongoDatabaseInterface {
		return &MongoDatabaseWrapper{
			db: &mongo.Database{},
		}
	}

	Convey("successful creation of new payable resources dao service", t, func() {
		cfg := &config.Config{
			MongoDBURL:                 dbUrl,
			Database:                   db,
			PayableResourcesCollection: "payable_resources",
		}

		pr := NewPayableResourcesDaoService(cfg)
		So(pr, ShouldNotBeNil)
	})

	Convey("successful creation of new account penalties dao service", t, func() {
		cfg := &config.Config{
			MongoDBURL:                 dbUrl,
			Database:                   db,
			AccountPenaltiesCollection: "account_penalties",
		}

		ap := NewAccountPenaltiesDaoService(cfg)
		So(ap, ShouldNotBeNil)
	})
}
