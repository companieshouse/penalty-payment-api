package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
)

// WriteJSON writes the interface as a json string with status of 200.
func WriteJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	WriteJSONWithStatus(w, r, data, http.StatusOK)
}

// WriteJSONWithStatus writes the interface as a json string with the supplied status.
func WriteJSONWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("error writing response: %v", err))
	}
}

// GetCompanyNumberFromVars returns the company number from the supplied request vars.
func GetCompanyNumberFromVars(vars map[string]string) (string, error) {
	companyNumber := vars["company_number"]
	log.Info("companyNumber: " + companyNumber)

	penaltyReference := vars["penalty_reference"]
	log.Info("penaltyReference: " + penaltyReference)

	if companyNumber == "" {
		return "", fmt.Errorf("company number not supplied")
	}

	return companyNumber, nil
}

// GetCompanyCodeFromVars returns the penalty reference from the supplied request vars.
func GetCompanyCodeFromVars(vars map[string]string) (string, error) {
	jsonData, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		log.Info("Error marshalling: ")
	}
	log.Info("JSON Data")
	log.Info(string(jsonData))

	var penaltyRef = "A00531369"
	log.Info("temp penRef: " + penaltyRef)

	return "LP", nil
}
