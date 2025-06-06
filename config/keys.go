package config

// ContextKey is used within the context api to store values
type ContextKey string

const (
	// CustomerCode is the key that stores the customer code
	CustomerCode = ContextKey("CustomerCode")
	// PayableResource is the key that stores the payable resource
	PayableResource = ContextKey("PayableResource")
)
