package config

// ContextKey is used within the context api to store values
type ContextKey string

const (
	// CompanyNumber is the key that stores the company number
	CompanyNumber = ContextKey("CompanyNumber")
	// PayableResource is the key that stores the payable resource
	PayableResource = ContextKey("PayableResource")
)
