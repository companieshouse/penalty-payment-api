package interceptors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/penalty-payment-api-core/constants"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/services"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func GetTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func createMockPayableResourceService(mockDAO *mocks.MockPayableResourceDaoService, cfg *config.Config) services.PayableResourceService {
	return services.PayableResourceService{
		DAO:    mockDAO,
		Config: cfg,
	}
}

// Function to create a PayableAuthenticationInterceptor with mock mongo DAO and a mock payment service
func createPayableAuthenticationInterceptorWithMockDAOAndService(controller *gomock.Controller, cfg *config.Config) PayableAuthenticationInterceptor {
	mockDAO := mocks.NewMockPayableResourceDaoService(controller)
	mockPayableResourceService := createMockPayableResourceService(mockDAO, cfg)
	return PayableAuthenticationInterceptor{
		Service: mockPayableResourceService,
	}
}

// Function to create a PayableAuthenticationInterceptor with the supplied payment service
func createPayableAuthenticationInterceptorWithMockService(PayableResourceService *services.PayableResourceService) PayableAuthenticationInterceptor {
	return PayableAuthenticationInterceptor{
		Service: *PayableResourceService,
	}
}

func TestUnitUserPaymentInterceptor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	cfg, _ := config.Get()

	Convey("No payment ID in request", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req.Header.Set("Eric-Identity", "authorised_identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "noroles")

		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockDAOAndService(mockCtrl, cfg)

		w := httptest.NewRecorder()
		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Invalid user details in context", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "authorised_identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "noroles")
		// The details have to be in a authUserDetails struct, so pass a different struct to fail
		authUserDetails := models.PayableResource{
			PayableRef: "test",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockDAOAndService(mockCtrl, cfg)

		w := httptest.NewRecorder()
		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Payable ref empty", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": ""})
		req.Header.Set("Eric-Identity", "authorised_identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "noroles")
		// The details have to be in a authUserDetails struct, so pass a different struct to fail
		authUserDetails := models.PayableResource{
			PayableRef: "test",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockDAOAndService(mockCtrl, cfg)

		w := httptest.NewRecorder()
		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("No authorised identity", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "authorised_identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "noroles")
		// Pass no ID (identity)
		authUserDetails := authentication.AuthUserDetails{}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockDAOAndService(mockCtrl, cfg)

		w := httptest.NewRecorder()
		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("Invalid identity header passed", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, _ := http.NewRequest("GET", path, nil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		authUserDetails := authentication.AuthUserDetails{
			ID: "INVALID",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		mockedGetAuthorisedIdentityType := func(r *http.Request) string {
			return "INVALID"
		}
		getAuthorisedIdentityType = mockedGetAuthorisedIdentityType

		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("Payment not found in DB", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "/admin/payment-lookup")
		authUserDetails := authentication.AuthUserDetails{
			ID: "identity",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		setMockedIdentityHeaderAsValid(authentication.Oauth2IdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(nil, nil)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error reading from DB", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "/admin/payment-lookup")
		authUserDetails := authentication.AuthUserDetails{
			ID: "identity",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(&models.PayableResourceDao{}, fmt.Errorf("error"))

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Happy path where user is creator", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "identity")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "noroles")
		authUserDetails := authentication.AuthUserDetails{
			ID: "identity",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		setMockedIdentityHeaderAsValid(authentication.Oauth2IdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		createdAt := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
				Data: models.PayableResourceDataDao{
					Etag:      "qwertyetag1234",
					CreatedAt: &createdAt,
					CreatedBy: models.CreatedByDao{
						ID: "identity",
					},
					Links: models.PayableResourceLinksDao{
						Self: "/company/12345678/penalties/payable/1234",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Happy path where user is admin and request is GET", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "admin")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "/admin/penalty-lookup")
		authUserDetails := authentication.AuthUserDetails{
			ID: "admin",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		setMockedIdentityHeaderAsValid(authentication.Oauth2IdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		createdAt := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
				Data: models.PayableResourceDataDao{
					Etag:      "qwertyetag1234",
					CreatedAt: &createdAt,
					CreatedBy: models.CreatedByDao{
						ID: "identity",
					},
					Links: models.PayableResourceLinksDao{
						Self: "/company/12345678/penalties/payable/1234",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Unauthorised where user is admin and request is POST", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("POST", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "admin")
		req.Header.Set("Eric-Identity-Type", authentication.Oauth2IdentityType)
		req.Header.Set("ERIC-Authorised-User", "test@test.com;test;user")
		req.Header.Set("ERIC-Authorised-Roles", "/admin/payment-lookup")
		authUserDetails := authentication.AuthUserDetails{
			ID: "admin",
		}
		ctx := context.WithValue(req.Context(), authentication.ContextKeyUserDetails, authUserDetails)

		setMockedIdentityHeaderAsValid(authentication.Oauth2IdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		createdAt := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
				Data: models.PayableResourceDataDao{
					Etag:      "qwertyetag1234",
					CreatedAt: &createdAt,
					CreatedBy: models.CreatedByDao{
						ID: "identity",
					},
					Links: models.PayableResourceLinksDao{
						Self: "/company/12345678/penalties/payable/1234",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req.WithContext(ctx))
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})

	Convey("Happy path where user has elevated privileges key accessing a non-creator resource", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "12345678", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "api_key")
		req.Header.Set("Eric-Identity-Type", authentication.APIKeyIdentityType)
		req.Header.Set("ERIC-Authorised-Key-Roles", "*")

		setMockedIdentityHeaderAsValid(authentication.APIKeyIdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		createdAt := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("12345678", "1234").Return(
			&models.PayableResourceDao{
				CustomerCode: "12345678",
				PayableRef:   "1234",
				Data: models.PayableResourceDataDao{
					Etag:      "qwertyetag1234",
					CreatedAt: &createdAt,
					CreatedBy: models.CreatedByDao{
						ID: "identity",
					},
					Links: models.PayableResourceLinksDao{
						Self: "/company/12345678/penalties/payable/1234",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Happy path where customer code is made uppercase", t, func() {
		path := fmt.Sprintf("/company/12345678/penalties/payable/%s", "1234")
		req, err := http.NewRequest("GET", path, nil)
		So(err, ShouldBeNil)
		req = mux.SetURLVars(req, map[string]string{"customer_code": "oc444555", "payable_ref": "1234"})
		req.Header.Set("Eric-Identity", "api_key")
		req.Header.Set("Eric-Identity-Type", authentication.APIKeyIdentityType)
		req.Header.Set("ERIC-Authorised-Key-Roles", "*")

		setMockedIdentityHeaderAsValid(authentication.APIKeyIdentityType)
		mockPrDaoSvc := mocks.NewMockPayableResourceDaoService(mockCtrl)
		mockPayableResourceSvc := createMockPayableResourceService(mockPrDaoSvc, cfg)
		payableAuthenticationInterceptor := createPayableAuthenticationInterceptorWithMockService(&mockPayableResourceSvc)

		txs := map[string]models.TransactionDao{
			"abcd": {Amount: 5},
		}
		createdAt := time.Now().Truncate(time.Millisecond)
		mockPrDaoSvc.EXPECT().GetPayableResource("OC444555", "1234").Return(
			&models.PayableResourceDao{
				CustomerCode: "OC444555",
				PayableRef:   "1234",
				Data: models.PayableResourceDataDao{
					Etag:      "qwertyetag1234",
					CreatedAt: &createdAt,
					CreatedBy: models.CreatedByDao{
						ID: "identity",
					},
					Links: models.PayableResourceLinksDao{
						Self: "/company/OC444555/penalties/payable/1234",
					},
					Transactions: txs,
					Payment: models.PaymentDao{
						Status: constants.Pending.String(),
						Amount: "5",
					},
				},
			},
			nil,
		)

		w := httptest.NewRecorder()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		test := payableAuthenticationInterceptor.PayableAuthenticationIntercept(GetTestHandler())
		test.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func setMockedIdentityHeaderAsValid(identityType string) {
	mockedGetAuthorisedIdentityType := func(r *http.Request) string {
		return identityType
	}
	getAuthorisedIdentityType = mockedGetAuthorisedIdentityType
}
