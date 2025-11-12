package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api-core/finance_config"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/handlers"
	. "github.com/smartystreets/goconvey/convey"
)

type mockConfigProvider struct {
	penaltyTypesConfig     []finance_config.FinancePenaltyTypeConfig
	payablePenaltiesConfig []finance_config.FinancePayablePenaltyConfig
}

func (m mockConfigProvider) GetPenaltyTypesConfig() []finance_config.FinancePenaltyTypeConfig {
	return m.penaltyTypesConfig
}

func (m mockConfigProvider) GetPayablePenaltiesConfig() []finance_config.FinancePayablePenaltyConfig {
	return m.payablePenaltiesConfig
}

type failingWriter struct{}

func (f failingWriter) Header() http.Header {
	return http.Header{}
}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("forced write error")
}

func (f failingWriter) WriteHeader(int) {}

func TestHandleConfiguration(t *testing.T) {
	now := time.Now()

	Convey("HandleConfiguration should respond correctly based on config type", t, func() {
		testCases := []struct {
			name           string
			queryParam     string
			expectedStatus int
			expectedType   string
			expectError    bool
			mockProvider   mockConfigProvider
		}{
			{
				name:           "Valid penaltyTypesConfig",
				queryParam:     "penaltyTypesConfig",
				expectedStatus: http.StatusOK,
				expectedType:   "penaltyTypesConfig",
				mockProvider: mockConfigProvider{
					penaltyTypesConfig: []finance_config.FinancePenaltyTypeConfig{
						{
							TransactionType:    "1",
							TransactionSubtype: "C1",
							Description:        "Late Filing Penalty",
							Disabled:           false,
							Reason:             "Late filing of accounts",
						},
					},
				},
			},
			{
				name:           "Valid payablePenaltiesConfig",
				queryParam:     "payablePenaltiesConfig",
				expectedStatus: http.StatusOK,
				expectedType:   "payablePenaltiesConfig",
				mockProvider: mockConfigProvider{
					payablePenaltiesConfig: []finance_config.FinancePayablePenaltyConfig{
						{
							EmailSend: &finance_config.EmailSendConfig{
								AppId:       "1",
								MessageType: "payment-reminder",
							},
							FinancePayment: &finance_config.FinancePaymentConfig{
								CompanyCode:     utils.LateFilingPenaltyCompanyCode,
								TransactionType: "1",
							},
							PaymentCost: &finance_config.PaymentCostConfig{
								ClassOfPayment: "penalty-lfp",
								Description:    "Late Filing Penalty",
								DescriptionId:  "Late Filing Penalty",
							},
							Penalty: &finance_config.PenaltyConfig{
								Reason:      "Late filing of accounts",
								EnabledFrom: now,
								EnabledTo:   now.Add(time.Hour),
							},
							Id:        "1",
							CreatedAt: now,
							CreatedBy: "admin",
							UpdatedAt: now,
							UpdatedBy: "admin",
						},
					},
				},
			},
			{
				name:           "Invalid config type",
				queryParam:     "invalidConfig",
				expectedStatus: http.StatusBadRequest,
				expectError:    true,
				mockProvider:   mockConfigProvider{},
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				req := httptest.NewRequest(http.MethodGet,
					"/penalty-payment-api/configuration?type="+tc.queryParam, nil)
				rr := httptest.NewRecorder()

				handler := handlers.HandleConfiguration(tc.mockProvider)
				handler.ServeHTTP(rr, req)

				So(rr.Code, ShouldEqual, tc.expectedStatus)

				if !tc.expectError {
					var resp struct {
						Type string      `json:"type"`
						Data interface{} `json:"data"`
					}
					err := json.NewDecoder(rr.Body).Decode(&resp)
					So(err, ShouldBeNil)
					So(resp.Type, ShouldEqual, tc.expectedType)
					So(resp.Data, ShouldNotBeNil)
				} else {
					So(rr.Body.String(), ShouldContainSubstring, "invalid configuration type supplied")
				}
			})
		}
	})

	Convey("HandleConfiguration should return 500 when JSON encoding fails", t, func() {
		req := httptest.NewRequest(http.MethodGet, "/penalty-payment-api/configuration?type=penaltyTypesConfig", nil)

		// Use failingWriter instead of httptest.ResponseRecorder
		fw := failingWriter{}

		mockProvider := mockConfigProvider{
			penaltyTypesConfig: []finance_config.FinancePenaltyTypeConfig{
				{
					TransactionType:    "1",
					TransactionSubtype: "C1",
					Reason:             "Late filing of accounts",
				},
			},
		}

		handler := handlers.HandleConfiguration(mockProvider)
		handler.ServeHTTP(fw, req)
	})
}
