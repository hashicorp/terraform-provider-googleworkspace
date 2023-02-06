// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func retryTimeDuration(ctx context.Context, duration time.Duration, retryFunc func() error) error {
	return resource.RetryContext(ctx, duration, func() *resource.RetryError {
		err := retryFunc()

		if err == nil {
			return nil
		}
		if IsNotConsistent(err) {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
}

func IsNotConsistent(err error) bool {
	errString, nErr := regexp.Compile("timed out while waiting")
	if nErr != nil {
		return false
	}
	matched := len(errString.FindAllStringSubmatch(err.Error(), 1)) > 0

	return matched
}

func isRetryableError(topErr error, customPredicates ...RetryErrorPredicateFunc) bool {
	if topErr == nil {
		return false
	}

	retryPredicates := append(
		// Global error retry predicates are registered in this default list.
		defaultErrorRetryPredicates,
		customPredicates...)

	// Check all wrapped errors for a retryable error status.
	isRetryable := false
	errwrap.Walk(topErr, func(werr error) {
		for _, pred := range retryPredicates {
			if predRetry, predReason := pred(werr); predRetry {
				log.Printf("[DEBUG] Dismissed an error as retryable. %s - %s", predReason, werr)
				isRetryable = true
				return
			}
		}
	})
	return isRetryable
}
