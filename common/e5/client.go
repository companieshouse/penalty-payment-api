package e5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/go-playground/validator.v9"

	"github.com/companieshouse/chs.go/log"
)

var (
	// ErrFailedToReadBody is a generic error when failing to parse a response body
	ErrFailedToReadBody = errors.New("failed reading the body of the response")
	// ErrE5BadRequest is a 400
	ErrE5BadRequest = errors.New("failed request to E5")
	// ErrE5InternalServer is anything in the 5xx
	ErrE5InternalServer = errors.New("got an internal server error from E5")
	// ErrE5NotFound is a 404
	ErrE5NotFound = errors.New("not found")
	// ErrUnexpectedServerError represents anything other than a 400, 404 or 500 - which would be something not
	// documented in their API
	ErrUnexpectedServerError = errors.New("unexpected server error")
)

// Action is the type that describes a payment call to E5
type Action string

const (
	// CreateAction signifies payment creation. This locks the customer account.
	CreateAction Action = "create"
	// AuthoriseAction signifies the payment has been authorised - but money not confirmed
	AuthoriseAction Action = "authorise"
	// ConfirmAction signifies money has been received. The customer account will not be unlocked
	ConfirmAction Action = "confirm"
	// TimeoutAction can be used to unlock the account following authorisation
	TimeoutAction Action = "timeout"
	// RejectAction will reject the payment altogether
	RejectAction Action = "reject"
)

// ClientInterface interface declares the Client finance system operations for AR Transactions and Payments
type ClientInterface interface {
	GetTransactions(input *GetTransactionsInput, requestId string) (*GetTransactionsResponse, error)
	CreatePayment(input *CreatePaymentInput, requestId string) error
	AuthorisePayment(input *AuthorisePaymentInput, requestId string) error
	ConfirmPayment(input *PaymentActionInput, requestId string) error
	TimeoutPayment(input *PaymentActionInput, requestId string) error
	RejectPayment(input *PaymentActionInput, requestId string) error
}

// Client interacts with the Client finance system
type Client struct {
	E5Username string
	E5BaseURL  string
}

// GetTransactions will return a list of transactions for a company
func (c *Client) GetTransactions(input *GetTransactionsInput, requestId string) (*GetTransactionsResponse, error) {
	err := c.validateInput(input)
	if err != nil {
		return nil, err
	}

	logContext := log.Data{"customer_code": input.CustomerCode}

	path := fmt.Sprintf("/arTransactions/%s", input.CustomerCode)
	qp := map[string]string{
		"companyCode": input.CompanyCode,
		"fromDate":    "1990-01-01",
	}

	// make the http request to E5
	resp, err := c.sendRequest(http.MethodGet, path, nil, qp, requestId)

	// deal with any http transport errors
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, err
	}

	defer closeResponseBody(resp, logContext)

	// determine if there are 4xx/5xx errors. an error here relates to a response parsing issue
	err = c.checkResponseForError(resp, requestId)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, err
	}

	out := &GetTransactionsResponse{
		Page:         Page{},
		Transactions: []Transaction{},
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, ErrFailedToReadBody
	}

	err = json.Unmarshal(b, out)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, ErrFailedToReadBody
	}

	return out, nil
}

// CreatePayment will create a new payment session in Client. This will lock the account in Client so no other modifications can
// happen until it is released by a confirm call or manually released in the Client portal.
func (c *Client) CreatePayment(input *CreatePaymentInput, requestId string) error {
	logContext := log.Data{
		"customer_code":  input.CustomerCode,
		"company_code":   input.CompanyCode,
		"payment_action": CreateAction,
		"payment_id":     input.PaymentID,
		"payment_value":  input.TotalValue,
		"transactions":   input.Transactions,
	}

	return c.doPaymentRequest(
		input,
		"/arTransactions/payment",
		logContext,
		"sending request to E5 Create Payment",
		"response received from E5 after creating a new payment",
		requestId,
	)
}

// AuthorisePayment will mark the payment as been authorised by the payment provider, but the money has not yet reached
// use yet. The customer account will remain locked.
func (c *Client) AuthorisePayment(input *AuthorisePaymentInput, requestId string) error {
	logContext := log.Data{
		"company_code":   input.CompanyCode,
		"payment_action": AuthoriseAction,
		"payment_id":     input.PaymentID,
	}

	return c.doPaymentRequest(
		input,
		"/arTransactions/payment/authorise",
		logContext,
		"sending request to E5 Authorise Payment",
		"response received from E5 after authorising a payment",
		requestId,
	)
}

// doPaymentRequest is a wrapper for the create and authorise endpoints
func (c *Client) doPaymentRequest(
	input interface{},
	path string,
	logContext log.Data,
	debugMsg string,
	infoMsg string,
	requestId string,
) error {
	err := c.validateInput(input)
	if err != nil {
		return err
	}

	body, err := json.Marshal(input)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return err
	}

	log.DebugC(requestId, debugMsg, logContext, log.Data{
		"input": input,
		"path":  path,
	})
	resp, err := c.sendRequest(http.MethodPost, path, bytes.NewReader(body), nil, requestId)

	// err here will be an http transport error rather than 4xx or 5xx responses
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return err
	}

	defer closeResponseBody(resp, logContext)

	log.InfoC(requestId, infoMsg, logContext, log.Data{
		"status": resp.StatusCode,
	})

	return c.checkResponseForError(resp, requestId)
}

// ConfirmPayment allocates the money in Client and unlocks the customer account
func (c *Client) ConfirmPayment(input *PaymentActionInput, requestId string) error {
	return c.doPaymentAction(ConfirmAction, input, requestId)
}

// TimeoutPayment will unlock the customer account
func (c *Client) TimeoutPayment(input *PaymentActionInput, requestId string) error {
	return c.doPaymentAction(TimeoutAction, input, requestId)
}

// RejectPayment will mark a payment as rejected and unlock the account.
func (c *Client) RejectPayment(input *PaymentActionInput, requestId string) error {
	return c.doPaymentAction(RejectAction, input, requestId)
}

// doPaymentAction is a wrapper for the confirm, reject and timeout endpoints
func (c *Client) doPaymentAction(action Action, input *PaymentActionInput, requestId string) error {
	err := c.validateInput(input)
	if err != nil {
		return err
	}

	logContext := log.Data{
		"company_code":   input.CompanyCode,
		"payment_action": action,
		"payment_id":     input.PaymentID,
	}

	body, err := json.Marshal(input)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return err
	}

	log.InfoC(requestId, "sending request to E5", logContext)

	path := fmt.Sprintf("/arTransactions/payment/%s", action)

	log.DebugC(requestId, "sending request to E5", logContext, log.Data{
		"input": input,
		"path":  path,
	})
	resp, err := c.sendRequest(http.MethodPost, path, bytes.NewReader(body), nil, requestId)

	// err here will be a http transport error rather than 4xx or 5xx responses
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return err
	}

	log.InfoC(requestId, "response received from E5", logContext)

	defer closeResponseBody(resp, logContext)

	return c.checkResponseForError(resp, requestId)
}

// generic function that inspects the http response and will return the response struct or an error if there was a
// problem reading and parsing the body
func (c *Client) checkResponseForError(r *http.Response, requestId string) error {

	if r.StatusCode == 200 {
		return nil
	}

	logContext := log.Data{
		"response_status": r.StatusCode,
	}

	// parse the error response and log all output
	e := &apiErrorResponse{}
	b, err := io.ReadAll(r.Body)

	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return ErrFailedToReadBody
	}

	err = json.Unmarshal(b, e)
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return ErrFailedToReadBody
	}

	d := log.Data{
		"http_status":   e.Code,
		"status":        e.Status,
		"message":       e.Message,
		"message_code":  e.MessageCode,
		"debug_message": e.DebugMessage,
		"errors":        e.SubErrorMap(),
	}

	log.ErrorC(requestId, errors.New("error response from E5"), d)

	switch r.StatusCode {
	case http.StatusBadRequest:
		return ErrE5BadRequest
	case http.StatusNotFound:
		return ErrE5NotFound
	case http.StatusInternalServerError:
		return ErrE5InternalServer
	default:
		return ErrUnexpectedServerError
	}
}

func (c *Client) validateInput(i interface{}) error {
	v := validator.New()
	return v.Struct(i)
}

// sendRequest will make a http request and unmarshal the response body into a struct
func (c *Client) sendRequest(method, path string, body io.Reader, queryParameters map[string]string, requestId string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.E5BaseURL, path)
	req, err := http.NewRequest(method, url, body)

	logContext := log.Data{"request_method": method, "path": path}
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// set query parameters
	qp := req.URL.Query()
	qp.Add("ADV_userName", c.E5Username)
	for k, v := range queryParameters {
		qp.Add(k, v)
	}

	req.URL.RawQuery = qp.Encode()

	resp, err := http.DefaultClient.Do(req)
	// any errors here are due to transport errors, not 4xx/5xx responses
	if err != nil {
		log.ErrorC(requestId, err, logContext)
		return nil, err
	}

	return resp, err
}

// NewClient will construct a new E5 client service struct that can be used to interact with the Client finance system
func NewClient(username, baseURL string) ClientInterface {
	return &Client{
		E5Username: username,
		E5BaseURL:  baseURL,
	}
}

func closeResponseBody(resp *http.Response, logContext log.Data) {
	err := resp.Body.Close()
	if err != nil {
		log.Error(err, logContext)
	}
}
