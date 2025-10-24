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
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/private"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var penaltyRefType = utils.LateFilingPenaltyRefType
var customerCode = "12345678"
var companyCode = "LP"
var penaltyDetailsMap = &config.PenaltyDetailsMap{
	Name: "penalty details",
	Details: map[string]config.PenaltyDetails{
		penaltyRefType: {
			Description:        "Late Filing Penalty",
			DescriptionId:      "late-filing-penalty",
			ClassOfPayment:     "penalty-lfp",
			ResourceKind:       "late-filing-penalty#late-filing-penalty",
			ProductType:        "late-filing-penalty",
			EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email",
			EmailMsgType:       "penalty_payment_received_email",
		},
	},
}

func TestUnitAccountPenalties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := types.AccountPenaltiesParams{
		PenaltyRefType:    penaltyRefType,
		CustomerCode:      customerCode,
		CompanyCode:       companyCode,
		PenaltyDetailsMap: penaltyDetailsMap,
		RequestId:         "",
	}
	createdAt := time.Now().Add(time.Minute * -10)
	staleCreatedAt := createdAt.Add(-24 * time.Hour)
	testE5TransactionResponse := &e5.GetTransactionsResponse{Transactions: make([]e5.Transaction, 1)}

	Convey("error when getting configuration data fails", t, func() {
		getConfig = func() (*config.Config, error) {
			return nil, errors.New("config error")
		}
		_, responseType, err := AccountPenalties(params)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	getConfig = config.Get
	cfg, _ := getConfig()
	cfg.AccountPenaltiesTTL = "24h"

	Convey("error when no transactions provided", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, nil)
		params.AccountPenaltiesDaoService = mockApDaoSvc
		_, responseType, err := AccountPenalties(params)
		So(err, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("penalties returned when valid transactions returned from cache (payableStatus = OPEN)", t, func() {
		accountPenalties := models.AccountPenaltiesDao{CreatedAt: &createdAt, ClosedAt: nil}
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(&accountPenalties, nil)

		testTransactionListResponse := utils.BuildTestTransactionListResponse(false, false, private.OpenPayableStatus, "LATE_FILING")
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return testTransactionListResponse, nil
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].PayableStatus, ShouldEqual, "OPEN")
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when valid transactions returned from cache (payableStatus = CLOSED_PENDING_ALLOCATION)", t, func() {
		accountPenalties := models.AccountPenaltiesDao{CreatedAt: &createdAt, ClosedAt: &createdAt}
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(&accountPenalties, nil)

		testTransactionListResponse := utils.BuildTestTransactionListResponse(false, true, private.ClosedPendingAllocationPayableStatus, "LATE_FILING")
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return testTransactionListResponse, nil
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].PayableStatus, ShouldEqual, "CLOSED_PENDING_ALLOCATION")
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned after a cache miss and account penalties created from e5 transaction (payableStatus = CLOSED)", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, errors.New("account penalties not found"))
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil)

		testTransactionListResponse := utils.BuildTestTransactionListResponse(true, true, private.ClosedPayableStatus, "SANCTIONS")
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return testTransactionListResponse, nil
		}
		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].PayableStatus, ShouldEqual, "CLOSED")
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("error creating account penalties cache entry from e5 transaction after cache miss", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, errors.New("account penalties not found"))
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(errors.New("error creating account penalties"))

		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return &models.TransactionListResponse{}, nil
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("penalties returned when stale transactions in cache but failed cache update", t, func() {
		accountPenalties := models.AccountPenaltiesDao{CreatedAt: &staleCreatedAt, ClosedAt: &staleCreatedAt}

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(&accountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any(), "").Return(errors.New("error updating account penalties"))

		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}

		params.AccountPenaltiesDaoService = mockPenaltiesService
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("cache updated and penalties returned when stale transactions in cache (PayableStatus = OPEN)", t, func() {
		// should default to 24h when empty string passed
		cfg.AccountPenaltiesTTL = ""
		accountPenalties := models.AccountPenaltiesDao{CreatedAt: &staleCreatedAt, ClosedAt: nil}

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(&accountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any(), "").Return(nil)

		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}
		testTransactionListResponse := utils.BuildTestTransactionListResponse(false, false, private.OpenPayableStatus, "SANCTIONS")
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return testTransactionListResponse, nil
		}

		params.AccountPenaltiesDaoService = mockPenaltiesService
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].PayableStatus, ShouldEqual, "OPEN")
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("cache updated and penalties returned when stale transactions in cache (PayableStatus = CLOSED)", t, func() {
		// should default to 24h when unparsable value passed
		cfg.AccountPenaltiesTTL = "24hhn"
		accountPenalties := models.AccountPenaltiesDao{CreatedAt: &staleCreatedAt, ClosedAt: &staleCreatedAt}

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(&accountPenalties, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any(), "").Return(nil)

		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}
		testTransactionListResponse := utils.BuildTestTransactionListResponse(false, true, private.ClosedPayableStatus, "LATE_FILING")
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return testTransactionListResponse, nil
		}

		params.AccountPenaltiesDaoService = mockPenaltiesService
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(listResponse.Items[0].PayableStatus, ShouldEqual, "CLOSED")
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("AccountPenalties not cached when E5 returns empty transactions for a given customer code", t, func() {

		mockPenaltiesService := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockPenaltiesService.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, nil)
		mockPenaltiesService.EXPECT().UpdateAccountPenalties(gomock.Any(), "").Return(nil).MaxTimes(0)
		mockPenaltiesService.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil).MaxTimes(0)

		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return &e5.GetTransactionsResponse{Transactions: make([]e5.Transaction, 0)}, nil
		}
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return &models.TransactionListResponse{}, nil
		}

		params.AccountPenaltiesDaoService = mockPenaltiesService
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldBeNil)
		So(listResponse, ShouldNotBeNil)
		So(len(listResponse.Items), ShouldEqual, 0)
		So(responseType, ShouldEqual, services.Success)
	})

	Convey("error when transactions cannot be found", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, nil)

		errGettingTransactions := errors.New("error getting transactions")
		getTransactions = func(customerCode string, companyCode string, client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return &e5.GetTransactionsResponse{}, errGettingTransactions
		}
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, penaltyRefType string,
			penaltyDetailsMap *config.PenaltyDetailsMap, cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return &models.TransactionListResponse{}, nil
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldEqual, errGettingTransactions)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})

	Convey("error when generating transaction list fails", t, func() {
		mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)
		mockApDaoSvc.EXPECT().GetAccountPenalties(customerCode, companyCode, "").Return(nil, nil)
		mockApDaoSvc.EXPECT().CreateAccountPenalties(gomock.Any(), "").Return(nil)

		errGeneratingTransactionList := errors.New("error generating transaction list from account penalties: [error generating etag]")
		getTransactions = func(customerCode string, companyCode string,
			client e5.ClientInterface, requestId string) (*e5.GetTransactionsResponse, error) {
			return testE5TransactionResponse, nil
		}
		generateTransactionList = func(accountPenalties *models.AccountPenaltiesDao, companyCode string, penaltyDetailsMap *config.PenaltyDetailsMap,
			cfg *config.Config, requestId string) (*models.TransactionListResponse, error) {
			return &models.TransactionListResponse{}, errors.New("error generating etag")
		}

		params.AccountPenaltiesDaoService = mockApDaoSvc
		listResponse, responseType, err := AccountPenalties(params)
		So(err, ShouldResemble, errGeneratingTransactionList)
		So(listResponse, ShouldBeNil)
		So(responseType, ShouldEqual, services.Error)
	})
}
