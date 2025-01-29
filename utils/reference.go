package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/companieshouse/penalty-payment-api-core/models"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
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

// GetCompanyNumberFromVars returns the company number from the supplied request vars.
func GetCompanyNumberFromVars(vars map[string]string) (string, error) {
	companyNumber := vars["company_number"]
	if len(companyNumber) == 0 {
		return "", fmt.Errorf("company number not supplied")
	}

	return strings.ToUpper(companyNumber), nil
}

// GetCompanyCode gets the company code from the prefix of the penalty reference
func GetCompanyCode(penaltyNumber string) (string, error) {
	if len(penaltyNumber) == 0 {
		return "", fmt.Errorf("penalty number not supplied")
	}

	penaltyPrefix := penaltyNumber[0]

	switch penaltyPrefix {
	case 'A':
		return "LP", nil
	case 'P':
		return "C1", nil
	default:
		return "", fmt.Errorf("invalid penalty number supplied")
	}
}

// GetCompanyCodeFromTransaction determines the penalty type by the penaltyNumber which is held in
// the first element of the transactions under the property TransactionID that is pulled back
func GetCompanyCodeFromTransaction(transactions []models.TransactionItem) (string, error) {
	if len(transactions) == 0 {
		return "", errors.New("no transactions found")
	}

	penaltyNumber := transactions[0].TransactionID
	penaltyType, err := GetCompanyCode(penaltyNumber)

	if err != nil {
		err = fmt.Errorf("error converting penalty number: [%v]", err)
		return "", err
	}
	return penaltyType, nil
}
