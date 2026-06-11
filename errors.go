package serront

import "fmt"

// Error is the single error type returned by every operation in this
// package. Code carries the API envelope's error.code (UPPER_SNAKE_CASE,
// e.g. NOT_FOUND, VALIDATION_ERROR, LIMIT_REACHED, UPGRADE_REQUIRED,
// INVALID_TRANSITION) or an SDK-side code (NETWORK_ERROR,
// INVALID_RESPONSE, DECODE_FAILED). Status is the HTTP status (0 for
// transport-level failures).
type Error struct {
	Status    int
	Code      string
	Message   string
	RequestID string
	Param     string
}

func (e *Error) Error() string {
	return fmt.Sprintf("serront: %s: %s", e.Code, e.Message)
}
