package dao

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

var companyCode = "LP"
var customerCode = "12345678"
var penaltyRef = "A1234567"

func TestMongoPayableResourceService_CreateAccountPenalties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	dao := &models.AccountPenaltiesDao{}

	svc := MongoAccountPenaltiesService{
		db:             mockDatabase,
		CollectionName: "account_penalties",
	}

	Convey("account penalties", t, func() {
		mockDatabase.EXPECT().Collection("account_penalties").Return(mockCollection)

		Convey("success when creating account penalty when no entry exists", func() {
			result := mongo.UpdateResult{
				MatchedCount:  0,
				UpsertedCount: 1,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.CreateAccountPenalties(dao)

			So(err, ShouldBeNil)
		})

		Convey("success when creating account penalty when entry already exists", func() {
			result := mongo.UpdateResult{
				MatchedCount: 1,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.CreateAccountPenalties(dao)

			So(err, ShouldBeNil)
		})

		Convey("error when creating account penalty", func() {
			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error creating payable resource"))

			err := svc.CreateAccountPenalties(dao)

			So(err, ShouldNotBeNil)
		})

	})

}

func TestMongoPayableResourceService_GetAccountPenalties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	svc := MongoAccountPenaltiesService{
		db:             mockDatabase,
		CollectionName: "account_penalties",
	}

	Convey("account penalties", t, func() {
		mockDatabase.EXPECT().Collection("account_penalties").Return(mockCollection)

		Convey("success when getting account penalty", func() {
			result := mongo.NewSingleResultFromDocument(bson.M{
				"customer_code": customerCode,
				"company_code":  companyCode,
			}, nil, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			resource, err := svc.GetAccountPenalties(companyCode, customerCode)

			So(err, ShouldBeNil)
			So(resource.CompanyCode, ShouldEqual, companyCode)
			So(resource.CustomerCode, ShouldEqual, customerCode)
		})

		Convey("error when no document returned when getting account penalty", func() {
			result := mongo.NewSingleResultFromDocument(bson.M{}, mongo.ErrNoDocuments, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			resource, err := svc.GetAccountPenalties(companyCode, customerCode)

			So(err, ShouldBeNil)
			So(resource, ShouldBeNil)
		})

		Convey("error when other error returned getting account penalty", func() {
			result := mongo.NewSingleResultFromDocument(nil, mongo.ErrInvalidIndexValue, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			_, err := svc.GetAccountPenalties(companyCode, customerCode)

			So(err, ShouldNotBeNil)
		})

	})

}

func TestMongoPayableResourceService_UpdateAccountPenaltyAsPaid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	svc := MongoAccountPenaltiesService{
		db:             mockDatabase,
		CollectionName: "account_penalties",
	}

	Convey("account penalties", t, func() {
		mockDatabase.EXPECT().Collection("account_penalties").Return(mockCollection)

		Convey("success when account penalty updated", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 1,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.UpdateAccountPenaltyAsPaid("", "", "")

			So(err, ShouldBeNil)
		})

		Convey("error when account penalty not updated due to DB error", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 0,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, errors.New("error updating as paid"))

			err := svc.UpdateAccountPenaltyAsPaid(companyCode, companyCode, penaltyRef)

			So(err, ShouldNotBeNil)
		})

		Convey("error when account penalty not updated", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 0,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.UpdateAccountPenaltyAsPaid(companyCode, companyCode, penaltyRef)

			So(err, ShouldNotBeNil)
		})

	})
}

func TestMongoPayableResourceService_UpdateAccountPenalties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	dao := &models.AccountPenaltiesDao{}

	svc := MongoAccountPenaltiesService{
		db:             mockDatabase,
		CollectionName: "account_penalties",
	}

	Convey("account penalties", t, func() {
		mockDatabase.EXPECT().Collection("account_penalties").Return(mockCollection)

		Convey("success when account penalties updated", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 1,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.UpdateAccountPenalties(dao)

			So(err, ShouldBeNil)
		})

		Convey("error when account penalty not updated due to DB error", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 0,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, errors.New("error updating penalties"))

			err := svc.UpdateAccountPenalties(dao)

			So(err, ShouldNotBeNil)
		})

		Convey("error when account penalties not updated", func() {

			result := mongo.UpdateResult{
				ModifiedCount: 0,
			}

			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(&result, nil)

			err := svc.UpdateAccountPenalties(dao)

			So(err, ShouldNotBeNil)
		})

	})
}

func TestMongoPayableResourceService_CreatePayableResource(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	dao := &models.PayableResourceDao{}

	svc := MongoPayableResourceService{
		db:             mockDatabase,
		CollectionName: "payable_resources",
	}

	Convey("payable resource", t, func() {
		mockDatabase.EXPECT().Collection("payable_resources").Return(mockCollection)

		Convey("success when creating payable resource", func() {
			mockCollection.EXPECT().InsertOne(gomock.Any(), gomock.Any()).Return(nil, nil)

			err := svc.CreatePayableResource(dao)

			So(err, ShouldBeNil)
		})

		Convey("error when creating payable resource", func() {
			mockCollection.EXPECT().InsertOne(gomock.Any(), gomock.Any()).Return(nil, errors.New("error creating payable resource"))

			err := svc.CreatePayableResource(dao)

			So(err, ShouldNotBeNil)
		})

	})

}

func TestMongoPayableResourceService_UpdatePaymentDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	dao := &models.PayableResourceDao{}

	svc := MongoPayableResourceService{
		db:             mockDatabase,
		CollectionName: "payable_resources",
	}

	Convey("payable resource", t, func() {
		mockDatabase.EXPECT().Collection("payable_resources").Return(mockCollection)

		Convey("success when updating payable resource", func() {
			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

			err := svc.UpdatePaymentDetails(dao)

			So(err, ShouldBeNil)
		})

		Convey("error when getting payable resource", func() {
			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

			err := svc.UpdatePaymentDetails(dao)

			So(err, ShouldNotBeNil)
		})

	})

}

//func TestMongoPayableResourceService_Shutdown(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)
//
//	svc := MongoPayableResourceService{
//		db:             mockDatabase,
//		CollectionName: "payable_resources",
//	}
//
//	Convey("shutdown", t, func() {
//		Convey("error no client", func() {
//
//			svc.Shutdown()
//		})
//
//		//Convey("error when creating payable resource", func() {
//		//	mockCollection.EXPECT().InsertOne(gomock.Any(), gomock.Any()).Return(nil, errors.New("error creating payable resource"))
//		//
//		//	svc.Shutdown()
//		//
//		//	So(err, ShouldNotBeNil)
//		//})
//
//	})
//
//}

func TestMongoPayableResourceService_GetPayableResource(t *testing.T) {
	customerCode := "1234568"
	payableRef := "1234568"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)

	svc := MongoPayableResourceService{
		db:             mockDatabase,
		CollectionName: "payable_resources",
	}

	payableRef = "1234568"

	Convey("payable resource", t, func() {
		mockDatabase.EXPECT().Collection("payable_resources").Return(mockCollection)

		Convey("success when getting payable resource", func() {
			result := mongo.NewSingleResultFromDocument(bson.M{
				"customer_code": customerCode,
				"payable_ref":   payableRef,
			}, nil, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			resource, err := svc.GetPayableResource(customerCode, payableRef)

			So(err, ShouldBeNil)
			So(resource.CustomerCode, ShouldEqual, customerCode)
			So(resource.PayableRef, ShouldEqual, payableRef)
		})

		Convey("error when getting payable resource", func() {
			result := mongo.NewSingleResultFromDocument(nil, nil, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			_, err := svc.GetPayableResource(customerCode, payableRef)

			So(err, ShouldNotBeNil)
		})

		Convey("error when getting no doc payable resource", func() {
			result := mongo.NewSingleResultFromDocument(bson.M{}, mongo.ErrNoDocuments, nil)

			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(result)

			resource, err := svc.GetPayableResource(customerCode, payableRef)

			So(err, ShouldBeNil)
			So(resource, ShouldBeNil)
		})

	})

}

//func TestMongoPayableResourceService_SaveE5Error(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockCollection := mocks.NewMockMongoCollectionInterface(ctrl)
//	mockDatabase := mocks.NewMockMongoDatabaseInterface(ctrl)
//
//	svc := MongoPayableResourceService{
//		db:             mockDatabase,
//		CollectionName: "payable_resources",
//	}
//
//	Convey("account penalties", t, func() {
//		mockDatabase.EXPECT().Collection("payable_resources").Return(mockCollection)
//
//		Convey("success when E5 error saved", func() {
//			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
//			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
//
//			err := svc.SaveE5Error(customerCode, penaltyRef, e5.CreateAction)
//
//			So(err, ShouldBeNil)
//		})
//
//		Convey("error E5 error not saved due to DB error", func() {
//			mockCollection.EXPECT().FindOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
//			mockCollection.EXPECT().UpdateOne(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error updating penalties"))
//
//			err := svc.SaveE5Error(customerCode, penaltyRef, e5.CreateAction)
//
//			So(err, ShouldNotBeNil)
//		})
//	})
//}
