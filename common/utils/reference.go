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
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
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
	rand.New(rand.NewSource(timeNow.UTC().UnixNano()))
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
func GetCompanyCode(penaltyRefType string) (string, error) {
	switch penaltyRefType {
	case LateFilingPenRef:
		return LateFilingPenaltyCompanyCode, nil
	case SanctionsPenRef:
		return SanctionsCompanyCode, nil
	case SanctionsRoePenRef:
		return SanctionsCompanyCode, nil
	default:
		return "", fmt.Errorf("invalid penalty reference type supplied")
	}
}

// GetCompanyCodeFromTransaction determines the penalty type by the penaltyReference which is held in
// the first element of the transactions under the property TransactionID that is pulled back
func GetCompanyCodeFromTransaction(transactions []models.TransactionItem) (string, error) {
	penaltyPrefix, err := getPrefix(transactions)
	if err != nil {
		return "", err
	}

	switch penaltyPrefix {
	case "A":
		return LateFilingPenaltyCompanyCode, nil
	case "P":
		return SanctionsCompanyCode, nil
	case "U":
		return SanctionsCompanyCode, nil
	default:
		return "", fmt.Errorf("error converting penalty reference")
	}
}

// GetPenaltyRefTypeFromTransaction determines the penalty reference type by the penaltyReference
// which is held in the first element of the transactions under the property TransactionID that is pulled back
func GetPenaltyRefTypeFromTransaction(transactions []models.TransactionItem) (string, error) {
	penaltyPrefix, err := getPrefix(transactions)
	if err != nil {
		return "", err
	}

	switch penaltyPrefix {
	case "A":
		return LateFilingPenRef, nil
	case "P":
		return SanctionsPenRef, nil
	case "U":
		return SanctionsRoePenRef, nil
	default:
		return "", fmt.Errorf("error converting penalty reference")
	}
}

func getPrefix(transactions []models.TransactionItem) (string, error) {
	if len(transactions) == 0 {
		return "", errors.New("no transactions found")
	}

	penaltyReference := transactions[0].PenaltyRef

	if len(penaltyReference) == 0 {
		return "", errors.New("no penalty reference found")
	}

	return penaltyReference[0:1], nil
}

const (
	LateFilingPenaltyCompanyCode = "LP"
	SanctionsCompanyCode         = "C1"
	LateFilingPenRef             = "LATE_FILING"
	SanctionsPenRef              = "SANCTIONS"
	SanctionsRoePenRef           = "SANCTIONS_ROE"
)
