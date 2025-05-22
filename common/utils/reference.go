package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/companieshouse/penalty-payment-api-core/models"
)

// GenerateReferenceNumber produces a random reference number in the format of [A-Z]{2}[0-9]{8}
func GenerateReferenceNumber() string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 2)
	for i := 0; i < 2; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b) + fmt.Sprintf("%08d", rand.Intn(99999999))
}

// GenerateEtag generates a random etag which is generated on every write action
func GenerateEtag() (string, error) {
	// Get a random number and the time in seconds and milliseconds
	timeNow := time.Now()
	rand.Seed(timeNow.UTC().UnixNano())
	randomNumber := fmt.Sprintf("%07d", rand.Intn(9999999))
	timeInMillis := strconv.FormatInt(timeNow.UnixNano()/int64(time.Millisecond), 10)
	timeInSeconds := strconv.FormatInt(timeNow.UnixNano()/int64(time.Second), 10)
	// Calculate a SHA-512 truncated digest
	shaDigest := sha512.New512_224()
	_, err := shaDigest.Write([]byte(randomNumber + timeInMillis + timeInSeconds))
	if err != nil {
		return "", fmt.Errorf("error writing sha digest: [%s]", err)
	}
	sha1Hash := hex.EncodeToString(shaDigest.Sum(nil))
	return sha1Hash, nil
}

// GetCustomerCodeFromVars returns the customer code from the supplied request vars.
func GetCustomerCodeFromVars(vars map[string]string) (string, error) {
	customerCode := vars["customer_code"]
	if len(customerCode) == 0 {
		return "", fmt.Errorf("customer code not supplied")
	}

	return strings.ToUpper(customerCode), nil
}

// GetCompanyCode gets the company code from the penalty reference type
func GetCompanyCode(penaltyReferenceType string) (string, error) {
	// If no penalty reference type is supplied then the request is coming in on the old url
	// so defaulting to LateFiling until agreement is made to update other services calling the api
	if len(penaltyReferenceType) == 0 {
		return LateFilingPenalty, nil
	}

	switch penaltyReferenceType {
	case "LATE_FILING":
		return LateFilingPenalty, nil
	case "SANCTIONS":
		return Sanctions, nil
	default:
		return "", fmt.Errorf("invalid penalty reference type supplied")
	}
}

// GetCompanyCodeFromTransaction determines the penalty type by the penaltyReference which is held in
// the first element of the transactions under the property TransactionID that is pulled back
func GetCompanyCodeFromTransaction(transactions []models.TransactionItem) (string, error) {
	if len(transactions) == 0 {
		return "", errors.New("no transactions found")
	}

	penaltyReference := transactions[0].PenaltyRef

	if len(penaltyReference) == 0 {
		return "", errors.New("no penalty reference found")
	}
	penaltyPrefix := penaltyReference[0]

	switch penaltyPrefix {
	case 'A':
		return LateFilingPenalty, nil
	case 'P':
		return Sanctions, nil
	default:
		return "", fmt.Errorf("error converting penalty reference")
	}
}

const (
	LateFilingPenalty = "LP"
	Sanctions         = "C1"
)
