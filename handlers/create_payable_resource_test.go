package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func serveCreatePayableResourceHandler(body []byte, payableResourceService dao.PayableResourceDaoService, payableRequest *models.PayableRequest) *httptest.ResponseRecorder {
	template := "/company/%s/penalties/payable"
	path := fmt.Sprintf(template, "10000024")
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler := CreatePayableResourceHandler(payableResourceService)
	handler.ServeHTTP(res, req.WithContext(testContext(payableRequest)))

	return res
}

func testContext(payableRequest *models.PayableRequest) context.Context {
	ctx := context.WithValue(context.Background(), config.RequestId, "abcd1234abcd1234abcd1234")
	if payableRequest != nil {
		ctx = context.WithValue(ctx, config.CreatePayableResource, *payableRequest)
	}

	return ctx
}

func TestUnitCreatePayableResourceHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
	customerCode := "10000024"
	requestId := "abcd1234abcd1234abcd1234"

	Convey("Error when retrieving payable resource request from context", t, func() {

		body, _ := json.Marshal(&models.PayableRequest{})

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, nil)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when failing to create payable resource", t, func() {
		mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any(), requestId).Return(errors.New("any error"))

		body, _ := json.Marshal(&models.PayableRequest{})

		payableRequest := models.PayableRequest{
			CustomerCode: customerCode,
			CreatedBy:    authentication.AuthUserDetails{},
			Transactions: make([]models.TransactionItem, 1),
		}

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, &payableRequest)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("successfully creating a payable resource", t, func() {
		mockPrDaoSvc.EXPECT().CreatePayableResource(gomock.Any(), requestId).Return(nil)

		body, _ := json.Marshal(&models.PayableRequest{})

		payableRequest := models.PayableRequest{
			CustomerCode: customerCode,
			CreatedBy: authentication.AuthUserDetails{
				Email:    "test@test.com",
				Forename: "test",
				Surname:  "test",
				ID:       "testId",
			},
			Transactions: []models.TransactionItem{
				{PenaltyRef: "A1234567", Amount: 150, MadeUpDate: "2017-02-28", Type: "penalty"},
			},
		}

		res := serveCreatePayableResourceHandler(body, mockPrDaoSvc, &payableRequest)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}
