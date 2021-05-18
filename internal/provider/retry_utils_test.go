package googleworkspace

import (
	"fmt"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsNotConsistent_retryable(t *testing.T) {
	err := fmt.Errorf("something timed out while waiting")
	isRetryable := IsNotConsistent(err)
	if !isRetryable {
		t.Errorf("inconsistent error not detected as temporarily unavailable")
	}
}

func TestIsNotConsistent_other(t *testing.T) {
	err := googleapi.Error{
		Code: 404,
		Body: "some text describing error",
	}
	isRetryable := IsNotConsistent(&err)
	if isRetryable {
		t.Errorf("404 error detected as inconsistency error")
	}
}

func TestIsTemporarilyUnavailableErrorCode_retryable(t *testing.T) {
	err := googleapi.Error{
		Code: 503,
		Body: "some text describing error",
	}
	isRetryable := IsTemporarilyUnavailable(&err)
	if !isRetryable {
		t.Errorf("503 error not detected as temporarily unavailable")
	}
}

func TestIsTemporarilyUnavailableErrorCode_other(t *testing.T) {
	err := googleapi.Error{
		Code: 404,
		Body: "Some unretryable issue",
	}
	isRetryable := IsTemporarilyUnavailable(&err)
	if isRetryable {
		t.Errorf("error incorrectly detected as temporarily unavailabl")
	}
}
