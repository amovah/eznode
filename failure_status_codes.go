package eznode

import "net/http"

// DefaultFailureStatusCodes is the default http status code which recognized as failure
var DefaultFailureStatusCodes = []int{
	http.StatusGone,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusTooManyRequests,
	http.StatusInternalServerError,
	http.StatusNotImplemented,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates,
	http.StatusInsufficientStorage,
	http.StatusLoopDetected,
	http.StatusNotExtended,
	http.StatusNetworkAuthenticationRequired,
	http.StatusPaymentRequired,
}
