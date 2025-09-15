package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func GetTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func createPayableResourceRequestValidator(mockDAO *mocks.MockAccountPenaltiesDaoService, penaltyDetailsMap *config.PenaltyDetailsMap, allowedTransactionsMap *models.AllowedTransactionMap) PayableResourceRequestValidator {
	return PayableResourceRequestValidator{
		PenaltyDetailsMap:      penaltyDetailsMap,
		AllowedTransactionsMap: allowedTransactionsMap,
		ApDaoService:           mockDAO,
	}
}

func TestUnitNonPayableRequests(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDAO := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
	defer mockCtrl.Finish()
	customerCode := "12345678"
	requestId := "abcd1234abcd1234abcd1234"

	Convey("requests to non payable endpoint are not processed", t, func() {

		path := fmt.Sprintf("/company/%s/penalties/", customerCode)
		req, err := http.NewRequest("GET", path, http.NoBody)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestUnitCreatePayableResourceRequestValidation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDAO := mocks.NewMockAccountPenaltiesDaoService(mockCtrl)
	defer mockCtrl.Finish()
	customerCode := "12345678"
	requestId := "abcd1234abcd1234abcd1234"

	Convey("Error decoding request body", t, func() {

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body := []byte{'{'}
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when user details not in context", t, func() {
		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when penalty reference type cannot be resolved", t, func() {

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: make([]models.TransactionItem, 1),
		})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		ctx := context.WithValue(context.Background(), authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
		ctx = context.WithValue(ctx, config.CustomerCode, customerCode)
		req = req.WithContext(ctx)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when company code cannot be resolved", t, func() {

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		ctx := context.WithValue(context.Background(), authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
		ctx = context.WithValue(ctx, config.CustomerCode, customerCode)
		req = req.WithContext(ctx)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		getCompanyCode = func(penaltyRefType string) (string, error) {
			return "", fmt.Errorf("error")
		}

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when getting payable penalty fails", t, func() {

		payablePenalty = func(params types.PayablePenaltyParams) (*models.TransactionItem, error) {
			return nil, fmt.Errorf("error")
		}

		ctx := context.WithValue(context.Background(), authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
		ctx = context.WithValue(ctx, config.CustomerCode, customerCode)

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		req = req.WithContext(ctx)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		getCompanyCode = utils.GetCompanyCode

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when create payable resource request validation fails", t, func() {

		payablePenalty = func(params types.PayablePenaltyParams) (*models.TransactionItem, error) {
			return &models.TransactionItem{}, nil
		}

		ctx := context.WithValue(context.Background(), authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
		ctx = context.WithValue(ctx, config.CustomerCode, customerCode)

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		req = req.WithContext(ctx)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		getCompanyCode = utils.GetCompanyCode

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Create payable resource request validation successful", t, func() {

		payablePenalty = func(params types.PayablePenaltyParams) (*models.TransactionItem, error) {
			return &models.TransactionItem{
				PenaltyRef: "A1234567",
				Amount:     150,
				MadeUpDate: "2017-02-28",
				Type:       "penalty",
				IsPaid:     false,
				IsDCA:      false,
				Reason:     "Late filing of accounts",
			}, nil
		}

		ctx := context.WithValue(context.Background(), authentication.ContextKeyUserDetails, authentication.AuthUserDetails{})
		ctx = context.WithValue(ctx, config.CustomerCode, customerCode)

		path := fmt.Sprintf("/company/%s/penalties/payable/", customerCode)
		body, _ := json.Marshal(&models.PayableRequest{
			CustomerCode: "10000024",
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		})
		req, err := http.NewRequest("POST", path, bytes.NewReader(body))
		req = req.WithContext(ctx)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-ID", requestId)

		getCompanyCode = utils.GetCompanyCode

		payableResourceValidator := createPayableResourceRequestValidator(mockDAO, &config.PenaltyDetailsMap{}, &models.AllowedTransactionMap{})

		w := httptest.NewRecorder()
		test := payableResourceValidator.PayableResourceValidate(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
