package dao

import (
	"testing"

	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatabase := &mongo.Database{}

	Convey("successful creation of new payable resources dao service", t, func() {
		mockMongoClientProvider := mocks.NewMockMongoClientProvider(ctrl)
		mockMongoClientProvider.EXPECT().Database("test").Return(mockDatabase)

		cfg := &config.Config{
			MongoDBURL:                 "mongodb://localhost:27017",
			Database:                   "test",
			PayableResourcesCollection: "payable_resources",
		}

		prDaoService := NewPayableResourcesDaoService(mockMongoClientProvider, cfg)
		So(prDaoService, ShouldNotBeNil)
	})

	Convey("successful creation of new account penalties dao service", t, func() {
		mockMongoClientProvider := mocks.NewMockMongoClientProvider(ctrl)
		mockMongoClientProvider.EXPECT().Database("test").Return(mockDatabase)

		cfg := &config.Config{
			MongoDBURL:                 "mongodb://localhost:27017",
			Database:                   "test",
			AccountPenaltiesCollection: "account_penalties",
		}

		apDaoService := NewAccountPenaltiesDaoService(mockMongoClientProvider, cfg)
		So(apDaoService, ShouldNotBeNil)
	})
}
