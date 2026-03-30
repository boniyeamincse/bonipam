package errors

// HTTPError is a consistent transport-level error contract.
type HTTPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
