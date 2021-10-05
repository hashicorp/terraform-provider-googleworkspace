package googleworkspace

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"

	"google.golang.org/api/googleapi"
)

type RetryErrorPredicateFunc func(error) (bool, string)

/** ADD GLOBAL ERROR RETRY PREDICATES HERE **/
// Retry predicates that shoud apply to all requests should be added here.
var defaultErrorRetryPredicates = []RetryErrorPredicateFunc{
	// Common network errors (usually wrapped by URL error)
	isNetworkTemporaryError,
	isNetworkTimeoutError,
	isIoEOFError,
	isConnectionResetNetworkError,

	// Common error codes
	isCommonRetryableErrorCode,
	isRateLimitExceeded,
}

/** END GLOBAL ERROR RETRY PREDICATES HERE **/

func isNetworkTemporaryError(err error) (bool, string) {
	if netErr, ok := err.(*net.OpError); ok && netErr.Temporary() {
		return true, "marked as timeout"
	}
	if urlerr, ok := err.(*url.Error); ok && urlerr.Temporary() {
		return true, "marked as timeout"
	}
	return false, ""
}

func isNetworkTimeoutError(err error) (bool, string) {
	if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		return true, "marked as timeout"
	}
	if urlerr, ok := err.(*url.Error); ok && urlerr.Timeout() {
		return true, "marked as timeout"
	}
	return false, ""
}

func isIoEOFError(err error) (bool, string) {
	if err == io.ErrUnexpectedEOF {
		return true, "Got unexpected EOF"
	}

	if urlerr, urlok := err.(*url.Error); urlok {
		wrappedErr := urlerr.Unwrap()
		if wrappedErr == io.ErrUnexpectedEOF {
			return true, "Got unexpected EOF"
		}
	}
	return false, ""
}

const connectionResetByPeerErr = ": connection reset by peer"

func isConnectionResetNetworkError(err error) (bool, string) {
	if strings.HasSuffix(err.Error(), connectionResetByPeerErr) {
		return true, fmt.Sprintf("reset connection error: %v", err)
	}
	return false, ""
}

// Retry on comon googleapi error codes for retryable errors.
// TODO(#5609): This may not need to be applied globally - figure out
// what retryable error codes apply to which API.
func isCommonRetryableErrorCode(err error) (bool, string) {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return false, ""
	}

	if gerr.Code == 500 || gerr.Code == 502 || gerr.Code == 503 {
		log.Printf("[DEBUG] Dismissed an error as retryable based on error code: %s", err)
		return true, fmt.Sprintf("Retryable error code %d", gerr.Code)
	}
	return false, ""
}

func isRateLimitExceeded(err error) (bool, string) {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return false, ""
	}

	if gerr.Code == 429 {
		log.Printf("[DEBUG] Dismissed an error as retryable based on error code: %s", err)
		return true, fmt.Sprintf("Retryable error code %d", gerr.Code)
	}

	if gerr.Code == 403 && strings.Contains(gerr.Error(), "quotaExceeded") {
		log.Printf("[DEBUG] Dismissed an error as retryable based on error code: %s", err)
		return true, fmt.Sprintf("Retryable error code %d", gerr.Code)
	}

	return false, ""
}
