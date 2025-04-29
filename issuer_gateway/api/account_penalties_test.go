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

var customerCode = "12345678"
var companyCode = "LP"
var penaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		utils.LateFilingPenalty: {
			Description:        "Late Filing Penalty",
			DescriptionId:      "late-filing-penalty",
			ClassOfPayment:     "penalty",
			ResourceKind:       "late-filing-penalty#late-filing-penalty",
			ProductType:        "late-filing-penalty",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		},
	},
}
var allowedTransactionMap = &models.AllowedTransactionMap{
	Types: map[string]map[string]bool{
		"1": {
			"EJ": true,
			"EK": true,
			"EL": true,
			"EU": true,
			"S1": true,
		},
	},
}
var page = e5.Page{
	Size:          6,
	TotalElements: 6,
	TotalPages:    1,
	Number:        0,
}
var e5TransactionsResponse = e5.GetTransactionsResponse{
	Page: page,
	Transactions: []e5.Transaction{
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A1234567",
			TransactionDate:      "2025-02-25",
			MadeUpDate:           "2025-02-12",
			Amount:               250,
			OutstandingAmount:    0,
			IsPaid:               true,
			TransactionType:      "1",
			TransactionSubType:   "EL",
			TypeDescription:      "Double DBL LTD E&W> 6 MNTHS   DLTWD     ",
			DueDate:              "2025-03-26",
			AccountStatus:        "CHS",
			DunningStatus:        "PEN1",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A0000001",
			TransactionDate:      "2020-07-21",
			MadeUpDate:           "2018-06-30",
			Amount:               3000,
			OutstandingAmount:    3000,
			IsPaid:               false,
			TransactionType:      "1",
			TransactionSubType:   "EL",
			TypeDescription:      "Double DBL LTD E&W> 6 MNTHS   DLTWD     ",
			DueDate:              "2020-07-21",
			AccountStatus:        "HLD",
			DunningStatus:        "PEN3        ",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "CF1",
			TransactionDate:      "2021-04-09",
			MadeUpDate:           "2018-06-30",
			Amount:               105,
			OutstandingAmount:    105,
			IsPaid:               false,
			TransactionType:      "5",
			TransactionSubType:   "19",
			TypeDescription:      "IREC             Payments from Debt Man ",
			DueDate:              "2021-04-09",
			AccountStatus:        "HLD",
			DunningStatus:        "            ",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "FC1",
			TransactionDate:      "2021-04-09",
			MadeUpDate:           "2018-06-30",
			Amount:               80,
			OutstandingAmount:    80,
			IsPaid:               false,
			TransactionType:      "5",
			TransactionSubType:   "19",
			TypeDescription:      "IREC             Payments from Debt Man ",
			DueDate:              "2021-04-09",
			AccountStatus:        "HLD",
			DunningStatus:        "            ",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A0000002",
			TransactionDate:      "2021-08-10",
			MadeUpDate:           "2019-06-30",
			Amount:               3000,
			OutstandingAmount:    0,
			IsPaid:               true,
			TransactionType:      "1",
			TransactionSubType:   "EL",
			TypeDescription:      "Double DBL LTD E&W> 6 MNTHS   DLTWD     ",
			DueDate:              "2021-08-10",
			AccountStatus:        "HLD",
			DunningStatus:        "PEN3        ",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A0000003",
			TransactionDate:      "2021-12-15",
			MadeUpDate:           "2020-06-26",
			Amount:               1500,
			OutstandingAmount:    1210,
			IsPaid:               false,
			TransactionType:      "1",
			TransactionSubType:   "EK",
			TypeDescription:      "Double DBL LTD E&W>6 MNTHS    DLTWC     ",
			DueDate:              "2021-12-15",
			AccountStatus:        "HLD",
			DunningStatus:        "PEN3        ",
		},
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A0000004",
			TransactionDate:      "2022-06-06",
			MadeUpDate:           "2021-06-26",
			Amount:               750,
			OutstandingAmount:    750,
			IsPaid:               false,
			TransactionType:      "1",
			TransactionSubType:   "EJ",
			TypeDescription:      "Double DBL LTD E&W>1<3 MNTH   DLTWB     ",
			DueDate:              "2022-06-06",
			AccountStatus:        "HLD",
			DunningStatus:        "PEN3        ",
		},
	},
}

var yesterday = time.Now().Add(-24 * time.Hour)
var yesterdayAt2Pm = time.Date(
	yesterday.Year(), yesterday.Month(), yesterday.Day(), 14, 0, 0, 0, time.Local)
var staleAccountPenalties = models.AccountPenaltiesDao{
	CustomerCode: "12345678",
	CompanyCode:  "LP",
	CreatedAt:    &yesterday,
	ClosedAt:     &yesterdayAt2Pm,
	AccountPenalties: []models.AccountPenaltiesDataDao{
		{
			CompanyCode:          "LP",
			LedgerCode:           "EW",
			CustomerCode:         "12345678",
			TransactionReference: "A1234567",
			TransactionDate:      "2025-02-25",
			MadeUpDate:           "2025-02-12",
			Amount:               250,
			OutstandingAmount:    0,
			IsPaid:               true,
			TransactionType:      "1",
			TransactionSubType:   "EL",
			TypeDescription:      "Double DBL LTD E&W> 6 MNTHS   DLTWD     ",
			DueDate:              "2025-03-26",
			AccountStatus:        "CHS",
			DunningStatus:        "PEN1",
		},
	},
}

func TestUnitAccountPenalties(t *testing.T) {
	cfg, _ := config.Get()
	cfg.AccountPenaltiesTTL = "24h"
	cfg.E5AllocationRoutineDuration = "4h"
	cfg.E5AllocationRoutineStartHour = 20
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("error when no transactions provided", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(nil, nil)
		_, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockApDaoSvc)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("Multiple payable late filing penalties, some with unpaid legal costs associated by made up date", t, func() {
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
		So(len(listResponse.Items), ShouldEqual, 7)
		So(responseType, ShouldEqual, services.Success)

		assertTransactionListItem(listResponse.Items[1], "A0000001", false, false,
			"2020-07-21", "2018-06-30", "2020-07-21",
			3000, 3000, "penalty", "Late filing of accounts", "CLOSED")
		assertTransactionListItem(listResponse.Items[2], "CF1", false, false,
			"2021-04-09", "2018-06-30", "2021-04-09",
			105, 105, "other", "Late filing of accounts", "CLOSED")
		assertTransactionListItem(listResponse.Items[3], "FC1", false, false,
			"2021-04-09", "2018-06-30", "2021-04-09",
			80, 80, "other", "Late filing of accounts", "CLOSED")
		assertTransactionListItem(listResponse.Items[4], "A0000002", true, false,
			"2021-08-10", "2019-06-30", "2021-08-10",
			3000, 0, "penalty", "Late filing of accounts", "CLOSED")
		assertTransactionListItem(listResponse.Items[5], "A0000003", false, false,
			"2021-12-15", "2020-06-26", "2021-12-15",
			1500, 1210, "penalty", "Late filing of accounts", "OPEN")
		assertTransactionListItem(listResponse.Items[6], "A0000004", false, false,
			"2022-06-06", "2021-06-26", "2022-06-06",
			750, 750, "penalty", "Late filing of accounts", "OPEN")
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

	Convey("penalties returned without cache update when stale transactions in cache and payment updated in cache but not in e5", t, func() {
		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&staleAccountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any()).MaxTimes(0)

		e5TransactionsResponse.Transactions[0].IsPaid = false
		getTransactions = func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockPenaltiesService)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("cache updated and penalties returned when stale transactions in cache and payment updated in e5 but not in cache", t, func() {
		staleAccountPenalties.AccountPenalties[0].IsPaid = false
		staleAccountPenalties.ClosedAt = nil

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
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("cache updated and penalties returned when stale transactions in cache and payment updated neither in e5 nor in cache", t, func() {
		staleAccountPenalties.AccountPenalties[0].IsPaid = false
		staleAccountPenalties.ClosedAt = nil

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&staleAccountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any()).Return(nil)

		e5TransactionsResponse.Transactions[0].IsPaid = false
		getTransactions = func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockPenaltiesService)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when stale transactions in cache but failed cache update", t, func() {
		staleAccountPenalties.AccountPenalties[0].IsPaid = true
		staleAccountPenalties.ClosedAt = nil

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode).Return(&staleAccountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any()).Return(errors.New("error updating account penalties"))

		e5TransactionsResponse.Transactions[0].IsPaid = true
		getTransactions = func(customerCode string, companyCode string,
			client *e5.Client) (*e5.GetTransactionsResponse, error) {
			return &e5TransactionsResponse, nil
		}

		listResponse, responseType, err := AccountPenalties(customerCode, companyCode, penaltyDetailsMap, allowedTransactionMap, mockPenaltiesService)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	// This test will fail between 8pm and 12.00am.
	Convey("cache updated and penalties returned when stale transactions in cache and payment updated in e5 and in cache", t, func() {
		staleAccountPenalties.AccountPenalties[0].IsPaid = true
		staleAccountPenalties.ClosedAt = &yesterday

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

func assertTransactionListItem(transactionListItem models.TransactionListItem, expectedID string, expectedIsPaid bool, expectedIsDCA bool,
	expectedDueDate string, expectedMadeUpDate string, expectedTransactionDate string,
	expectedOriginalAmount float64, expectedOutstandingAmount float64, expectedType string, expectedReason string, expectedPayableStatus string) {
	expectedTransactionListItem := models.TransactionListItem{
		ID:              expectedID,
		Etag:            transactionListItem.Etag,
		Kind:            "late-filing-penalty#late-filing-penalty",
		IsPaid:          expectedIsPaid,
		IsDCA:           expectedIsDCA,
		DueDate:         expectedDueDate,
		MadeUpDate:      expectedMadeUpDate,
		TransactionDate: expectedTransactionDate,
		OriginalAmount:  expectedOriginalAmount,
		Outstanding:     expectedOutstandingAmount,
		Type:            expectedType,
		Reason:          expectedReason,
		PayableStatus:   expectedPayableStatus,
	}
	So(transactionListItem, ShouldResemble, expectedTransactionListItem)
}
