package eznode

import "net/http"

var DefaultFailureStatusCodes = []int{
	http.StatusBadRequest,
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
}
