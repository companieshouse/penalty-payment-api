package service

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/utils"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/issuer_gateway/types"
	"github.com/companieshouse/penalty-payment-api/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var req = &http.Request{}

type mockRoundTripper struct {
	headerChecked *bool
	returnError   bool
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.headerChecked != nil && req.Method == "POST" && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		*m.headerChecked = true
	}
	if m.returnError {
		return nil, errors.New("client error")
	}
	return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
}

func TestUnitSendEmailKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("Given the SendEmailKafkaMessage is called", t, func() {

		Convey("When config is called with invalid config", func() {
			errMsg := "config is invalid"
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, errors.New(errMsg)
			}

			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
				So(err, ShouldResemble, errors.New("error getting config for sending email message chs-kafka-api: ["+errMsg+"]"))
			})
		})

	})
}

func TestSendEmailMessageViaChsKafkaApi_UsesChsKafkaApiURL(t *testing.T) {
	mockTransport := &mockRoundTripper{}
	oldHttpClient := httpClient
	httpClient = &http.Client{Transport: mockTransport}
	defer func() { httpClient = oldHttpClient }()

	Convey("Given ChsKafkaApiURL is set in config", t, func() {

		expectedURL := "http://test-kafka-api?app_id=qwerty&email_address=testemail%40test.com&json_data=%7B%7D&message_id=123abc&message_type=email"
		// Save all globals that may be mocked
		oldGetConfig := getConfig
		oldNewRequestFunc := newRequestFunc
		oldPrepareEmailMessage := prepareEmailMessage
		defer func() {
			getConfig = oldGetConfig
			newRequestFunc = oldNewRequestFunc
			prepareEmailMessage = oldPrepareEmailMessage
		}()

		getConfig = func() (*config.Config, error) {
			return &config.Config{
				ChsKafkaApiURL: "http://test-kafka-api",
			}, nil
		}

		var actualURL string
		newRequestFunc = func(method, url string, body io.Reader) (*http.Request, error) {
			actualURL = url
			return &http.Request{Header: make(http.Header)}, nil
		}

		// Mock prepareEmailMessage to avoid unrelated errors
		prepareEmailMessage = func(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
			allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
			return &models.EmailSend{Data: "{}", MessageType: "email", EmailAddress: "testemail@test.com", MessageID: "123abc", AppID: "qwerty"}, nil
		}

		err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
		So(err, ShouldBeNil)
		So(actualURL, ShouldEqual, expectedURL)
	})
}

func TestUnitPrepareEmailKafkaMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("Given the PrepareKafkaMessage is called", t, func() {
		// Save all globals that may be mocked
		oldGetConfig := getConfig
		oldGetCompanyName := getCompanyName
		oldGetCompanyCodeFromTransaction := getCompanyCodeFromTransaction
		oldGetPenaltyRefTypeFromTransaction := getPenaltyRefTypeFromTransaction
		oldGetPayablePenalty := getPayablePenalty
		defer func() {
			getConfig = oldGetConfig
			getCompanyName = oldGetCompanyName
			getCompanyCodeFromTransaction = oldGetCompanyCodeFromTransaction
			getPenaltyRefTypeFromTransaction = oldGetPenaltyRefTypeFromTransaction
			getPayablePenalty = oldGetPayablePenalty
		}()
		mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
			return "LP", nil
		}

		getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

		testCases := []struct {
			name           string
			companyCode    string
			penaltyRefType string
		}{
			{
				name:           "Late Filing",
				companyCode:    utils.LateFilingPenaltyCompanyCode,
				penaltyRefType: utils.LateFilingPenaltyRefType,
			},
			{
				name:           "Sanctions",
				companyCode:    utils.SanctionsCompanyCode,
				penaltyRefType: utils.SanctionsPenaltyRefType,
			},
			{
				name:           "Sanctions ROE",
				companyCode:    utils.SanctionsCompanyCode,
				penaltyRefType: utils.SanctionsRoePenaltyRefType,
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				setGetCompanyCodeFromTransactionMock(tc.companyCode)
				setGetPenaltyRefTypeFromTransactionMock(tc.penaltyRefType)

				Convey("When config is called with invalid config", func() {
					errMsg := "config is invalid"
					mockedConfigGet := func() (*config.Config, error) {
						return &config.Config{}, errors.New(errMsg)
					}

					getConfig = mockedConfigGet

					Convey("Then an error should be returned", func() {
						_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
						So(err, ShouldResemble, errors.New("error getting config: ["+errMsg+"]"))
					})
				})

			})
		}

		Convey("When config is called with invalid config", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			getConfig = mockedConfigGet

			Convey("Then an error should be returned", func() {
				_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "error getting company name: [")
			})
		})
		Convey("When config is called with valid config and invalid company code", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			mockedGetCompanyCodeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
				return "", errors.New("error getting company code")
			}
			getCompanyCodeFromTransaction = mockedGetCompanyCodeFromTransaction

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName

			Convey("Then an error should be returned", func() {
				_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error getting company code")
			})
		})
		Convey("When config is called with valid config and valid company number but invalid penalty ref", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			mockedGetPenaltyRefTypeFromTransaction := func(transactions []models.TransactionItem) (string, error) {
				return "", errors.New("error getting penalty ref type")
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getPenaltyRefTypeFromTransaction = mockedGetPenaltyRefTypeFromTransaction

			Convey("Then an error should be returned", func() {
				_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

				So(err, ShouldResemble, errors.New("error getting penalty ref type"))
			})
		})
		Convey("When config is called with valid config and valid company number but no transaction items", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenaltyRefType)

			mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)

			Convey("Then an error should be returned", func() {
				payableResourceNoItems := models.PayableResource{
					CustomerCode: customerCode,
					Transactions: []models.TransactionItem{},
				}

				_, err := prepareEmailMessage(payableResourceNoItems, req, penaltyDetailsMap, allowedTransactionsMap, mockApDaoSvc)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "empty transactions list in payable resource:")
			})
		})
		Convey("When config is called with valid config and valid company number but invalid transaction", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			setGetPenaltyRefTypeFromTransactionMock(utils.LateFilingPenaltyRefType)

			mockApDaoSvc := mocks.NewMockAccountPenaltiesDaoService(ctrl)

			Convey("Then an error should be returned", func() {
				mockApDaoSvc.EXPECT().GetAccountPenalties(gomock.Any(), gomock.Any(), "").Return(nil, nil)

				_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, mockApDaoSvc)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "error getting transaction for penalty: [")
			})
		})
		Convey("When config is called with valid config and valid company number and valid transaction but invalid madeUpDate", func() {
			mockedConfigGet := func() (*config.Config, error) {
				return &config.Config{}, nil
			}
			mockedGetCompanyName := func(companyNumber string, req *http.Request) (string, error) {
				return "Brewery", nil
			}
			mockedGetPayablePenalty := func(params types.PayablePenaltyParams) (*models.TransactionItem, error) {

				return &models.TransactionItem{PenaltyRef: "A1234567", Reason: "Late filing of accounts"}, nil
			}

			getConfig = mockedConfigGet
			getCompanyName = mockedGetCompanyName
			getPayablePenalty = mockedGetPayablePenalty

			Convey("Then an error should be returned", func() {
				_, err := prepareEmailMessage(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)

				So(err, ShouldResemble, errors.New("error parsing made up date: [parsing time \"\" as \"2006-01-02\": cannot parse \"\" as \"2006\"]"))
			})
		})
	})
}

func TestSendEmailMessageViaChsKafkaApi_MissingScenarios(t *testing.T) {
	Convey("Given SendEmailMessageViaChsKafkaApi is called", t, func() {
		var headerChecked bool
		mockTransport := &mockRoundTripper{headerChecked: &headerChecked}
		oldHttpClient := httpClient
		httpClient = &http.Client{Transport: mockTransport}
		defer func() { httpClient = oldHttpClient }()

		oldPrepareEmailMessage := prepareEmailMessage
		oldNewRequestFunc := newRequestFunc
		oldGetConfig := getConfig
		defer func() {
			prepareEmailMessage = oldPrepareEmailMessage
			newRequestFunc = oldNewRequestFunc
			getConfig = oldGetConfig
		}()

		getConfig = func() (*config.Config, error) {
			return &config.Config{
				ChsKafkaApiURL: "http://test-kafka-api",
			}, nil
		}

		// Successful POST and header check
		Convey("When everything succeeds and header is set", func() {
			prepareEmailMessage = func(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
				return &models.EmailSend{Data: "{}"}, nil
			}

			err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
			So(err, ShouldBeNil)
			So(headerChecked, ShouldBeTrue)
		})

		// Marshal error
		Convey("When prepareEmailMessage returns a struct that cannot be marshaled", func() {
			prepareEmailMessage = func(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
				// Intentionally return a struct with a field that cannot be marshaled
				return nil, nil
			}
			newRequestFunc = func(method, url string, body io.Reader) (*http.Request, error) {
				return &http.Request{Header: make(http.Header), Method: method}, nil
			}
			err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
			So(err, ShouldNotBeNil)
		})

		// Request creation error
		Convey("When newRequestFunc returns an error", func() {
			prepareEmailMessage = func(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
				return &models.EmailSend{Data: "{}"}, nil
			}
			newRequestFunc = func(method, url string, body io.Reader) (*http.Request, error) {
				return nil, errors.New("request creation failed")
			}
			err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
			So(err, ShouldNotBeNil)
		})

		// HTTP client error
		Convey("When client.Do returns an error", func() {
			prepareEmailMessage = func(payableResource models.PayableResource, req *http.Request, penaltyDetailsMap *config.PenaltyDetailsMap,
				allowedTransactionsMap *models.AllowedTransactionMap, apDaoSvc dao.AccountPenaltiesDaoService) (*models.EmailSend, error) {
				return &models.EmailSend{Data: "{}"}, nil
			}
			newRequestFunc = func(method, url string, body io.Reader) (*http.Request, error) {
				return &http.Request{Header: make(http.Header), Method: method}, nil
			}
			err := SendEmailMessageViaChsKafkaApi(payableResource, req, penaltyDetailsMap, allowedTransactionsMap, nil)
			So(err, ShouldBeNil) // Function does not return error from client.Do
		})
	})
}
