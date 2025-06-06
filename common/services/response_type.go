package services

// ResponseType enumerates the types of authentication supported
type ResponseType int

const (
	// InvalidData response
	InvalidData ResponseType = iota

	// Error response
	Error

	// Forbidden response
	Forbidden

	// NotFound response
	NotFound

	// Success response
	Success
)

var vals = [...]string{
	"invalid-data",
	"error",
	"forbidden",
	"not-found",
	"success",
}

// String representation of `ResponseType`
func (a ResponseType) String() string {
	return vals[a]
}
