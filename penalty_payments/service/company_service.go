package service

import (
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
)

// GetCompanyName will attempt to get the company name from the CompanyProfileAPI.
func GetCompanyName(customerCode string, req *http.Request) (string, error) {

	api, err := manager.GetSDK(req)
	if err != nil {
		log.ErrorR(req, err, log.Data{"customer_code": customerCode})
		return "", err
	}

	companyProfile, err := api.Profile.Get(customerCode).Do()
	if err != nil {
		log.ErrorR(req, err, log.Data{"customer_code": customerCode})
		return "", err
	}

	return companyProfile.CompanyName, nil
}
