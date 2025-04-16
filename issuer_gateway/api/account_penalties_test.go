package api

import (
	"errors"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var customerCode = "NI123456"
var companyCode = "LP"
var penaltyDetailsMap = &config.PenaltyDetailsMap{}
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EU": true,
		},
	},
}
var transaction = e5.Transaction{
	CompanyCode:     "LP",
	TransactionType: "EU",
}
var page = e5.Page{
	Size:          1,
	TotalElements: 1,
	TotalPages:    1,
	Number:        1,
}
var e5TransactionsResponse = e5.GetTransactionsResponse{
	Page: page,
	Transactions: []e5.Transaction{
		0: transaction,
	},
}

var yesterday = time.Now().Add(-24 * time.Hour)
var staleAccountPenalties = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  utils.LateFilingPenalty,
	CreatedAt:    &yesterday,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		{
			CompanyCode:          utils.LateFilingPenalty,
			LedgerCode:           "E1",
			CustomerCode:         "12345678",
			TransactionReference: "P1234567",
			TransactionDate:      "2025-02-25",
			MadeUpDate:           "2025-02-12",
			Amount:               250,
			OutstandingAmount:    250,
			IsPaid:               false,
			TransactionType:      "1",
			TransactionSubType:   "S1",
			TypeDescription:      "CS01",
			DueDate:              "2025-03-26",
			AccountStatus:        "CHS",
			DunningStatus:        "PEN1",
		},
	},
}

func TestUnitAccountPenalties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("error when no transactions provided", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)
		_, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("penalties returned when valid transactions", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(nil)

		mockedGetTransactions := func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when valid transactions but error creating account penalties cache entry", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(errors.New("error creating account penalties"))

		mockedGetTransactions := func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when valid transactions returned from cache", t, func() {
		today := time.Now().Add(time.Hour * 1)

		accountPenalties := models.AccountPenaltiesDao{
			CustomerCode: "12345678",
			CompanyCode:  utils.Sanctions,
			CreatedAt:    &today,
			AccountPenalties: []models.AccountPenaltiesDataDao{
				{
					CompanyCode:          utils.Sanctions,
					LedgerCode:           "E1",
					CustomerCode:         "12345678",
					TransactionReference: "P1234567",
					TransactionDate:      "2025-02-25",
					MadeUpDate:           "2025-02-12",
					Amount:               250,
					OutstandingAmount:    250,
					IsPaid:               false,
					TransactionType:      "1",
					TransactionSubType:   "S1",
					TypeDescription:      "CS01",
					DueDate:              "2025-03-26",
					AccountStatus:        "CHS",
					DunningStatus:        "PEN1",
				},
			},
		}

		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&accountPenalties, nil)

		mockedGetTransactions := func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when stale transactions in cache but failed cache update", t, func() {

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&staleAccountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any()).Return(errors.New("error updating account penalties"))

		getTransactions = func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockPenaltiesService)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when when stale transactions in cache and successful cache update", t, func() {
		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&staleAccountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any()).Return(nil)

		e5TransactionsResponse.Transactions[0].IsPaid = true
		getTransactions = func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockPenaltiesService)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].IsPaid, ShouldEqual, true)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("error when transactions cannot be found", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)

		errGettingTransactions := errors.New("error getting transactions")
		mockedGetTransactions := func(customerCode string, companyCode string, client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5.GetTransactionsResponse{}, errGettingTransactions
		}

		getTransactions = mockedGetTransactions

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldEqual, errGettingTransactions)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("error when generating transaction list fails", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any()).Return(nil)

		errGeneratingTransactionList := errors.New("error generating transaction list from account penalties: [error generating etag]")
		payableTransactionList := models.TransactionListResponse{}
		mockedGetTransactions := func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}
		mockedGenerateTransactionList := func(accountPenalties *models.AccountPenaltiesDao, companyCode string,
			penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) (*models.TransactionListResponse, error) {
			return &payableTransactionList, errors.New("error generating etag")
		}

		getTransactions = mockedGetTransactions
		generateTransactionList = mockedGenerateTransactionList

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldResemble, errGeneratingTransactionList)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("error when getConfig fails", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)

		errGettingConfig := errors.New("error getting config")
		mockedGetConfig := func() (*config.Config, error) {
			return nil, errGettingConfig
		}

		getConfig = mockedGetConfig

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldBeNil)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})
}
