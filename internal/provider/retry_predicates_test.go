// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"strconv"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsCommonRetryableErrorCode_retryableErrorCode(t *testing.T) {
	codes := []int{500, 502, 503}
	for _, code := range codes {
		code := code
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			err := googleapi.Error{
				Code: code,
				Body: "some text describing error",
			}
			isRetryable, _ := isCommonRetryableErrorCode(&err)
			if !isRetryable {
				t.Errorf("Error not detected as retryable")
			}
		})
	}
}

func TestIsCommonRetryableErrorCode_otherError(t *testing.T) {
	err := googleapi.Error{
		Code: 404,
		Body: "Some unretryable issue",
	}
	isRetryable, _ := isCommonRetryableErrorCode(&err)
	if isRetryable {
		t.Errorf("Error incorrectly detected as retryable")
	}
}

func TestIsOperationReadQuotaError_quotaExceeded(t *testing.T) {
	err := googleapi.Error{
		Code: 403,
		Body: "Request rate higher than configured., quotaExceeded",
	}
	isRetryable, _ := isRateLimitExceeded(&err)
	if !isRetryable {
		t.Errorf("Error not detected as retryable")
	}
}

func TestIsOperationReadQuotaError_rateLimitExceeded(t *testing.T) {
	err := googleapi.Error{
		Code: 429,
		Body: "Rate Limit Exceeded., rateLimitExceeded",
	}
	isRetryable, _ := isRateLimitExceeded(&err)
	if !isRetryable {
		t.Errorf("Error not detected as retryable")
	}
}

func TestGoogle404Error(t *testing.T) {
	gerr := googleapi.Error{
		Code:    404,
		Message: "notfound",
	}
	err := &gerr

	expected := true

	if isNotFound(err) != expected {
		t.Error("Failed: The error should have been detected as 404")
	}
}

func TestGoogleNot404Error(t *testing.T) {
	gerr := googleapi.Error{
		Code:    200,
		Message: "notfound",
	}
	err := &gerr

	expected := false

	if isNotFound(err) != expected {
		t.Error("Failed: The error was detected as a 404 but should not have been")
	}
}
