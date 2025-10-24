package utils

import "github.com/companieshouse/penalty-payment-api-core/models"

func BuildTestTransactionListResponse(isDCA, isPaid bool, payableStatus, penaltyReferenceType string) *models.TransactionListResponse {
	transactionListItem := BuildTestTransactionListItem(isDCA, isPaid, payableStatus, penaltyReferenceType)

	return &models.TransactionListResponse{
		Etag:  "abc1234",
		Items: []models.TransactionListItem{transactionListItem},
	}
}

func BuildTestTransactionListItem(isDCA, isPaid bool, payableStatus, penaltyReferenceType string) models.TransactionListItem {
	id, kind, penaltyType, reason := getPenaltyRefTypeSpecificParameters(penaltyReferenceType)

	outstandingAmount := 250.0
	if isPaid {
		outstandingAmount = 0.0
	}
	return models.TransactionListItem{
		ID:              id,
		Etag:            "abc1234",
		Kind:            kind,
		IsPaid:          isPaid,
		IsDCA:           isDCA,
		DueDate:         "2025-03-26",
		MadeUpDate:      "2025-02-12",
		TransactionDate: "2025-02-25",
		OriginalAmount:  250,
		Outstanding:     outstandingAmount,
		Type:            penaltyType,
		Reason:          reason,
		PayableStatus:   payableStatus,
	}
}

func BuildTestTransactionItem(isDCA, isPaid bool, penaltyReferenceType string, amount float64) models.TransactionItem {
	penaltyRef, _, penaltyType, reason := getPenaltyRefTypeSpecificParameters(penaltyReferenceType)

	return models.TransactionItem{
		PenaltyRef: penaltyRef,
		IsPaid:     isPaid,
		IsDCA:      isDCA,
		MadeUpDate: "2025-02-12",
		Amount:     amount,
		Type:       penaltyType,
		Reason:     reason,
	}
}

func getPenaltyRefTypeSpecificParameters(penaltyReferenceType string) (string, string, string, string) {
	id := "Z1234567"
	kind := "other#other"
	penaltyType := "other"
	reason := "penalty"

	if penaltyReferenceType == "SANCTIONS" {
		id = "P1234567"
		kind = "penalty#sanctions"
		penaltyType = "penalty"
		reason = "Failure to file a confirmation statement"
	} else if penaltyReferenceType == "SANCTIONS_ROE" {
		id = "U1234567"
		kind = "penalty#sanctions"
		penaltyType = "penalty"
		reason = "Failure to update the Register of Overseas Entities"
	} else if penaltyReferenceType == "LATE_FILING" {
		id = "A1234567"
		kind = "late-filing-penalty#late-filing-penalty"
		penaltyType = "penalty"
		reason = "Late filing of accounts"
	}

	return id, kind, penaltyType, reason
}
