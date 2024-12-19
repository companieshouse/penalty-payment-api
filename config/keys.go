package config

// ContextKey is used within the context api to store values
type ContextKey string

const (
	// CompanyDetails is the key that stores the company number and company code
	CompanyDetails = ContextKey("CompanyDetails")
	// PayableResource is the key that stores the payable resource
	PayableResource = ContextKey("PayableResource")
)
