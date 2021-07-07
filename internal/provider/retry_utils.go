package googleworkspace

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"google.golang.org/api/googleapi"
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

		if IsTemporarilyUnavailable(err) {
			return resource.RetryableError(err)
		}

		if IsRateLimitExceeded(err) {
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

func IsTemporarilyUnavailable(err error) bool {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	if gerr.Code == 503 {
		log.Printf("[DEBUG] Dismissed an error as retryable based on error code: %s", err)
		return true
	}
	return false

}

func IsRateLimitExceeded(err error) bool {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	if gerr.Code == 429 {
		log.Printf("[DEBUG] Dismissed an error as retryable based on error code: %s", err)
		return true
	}
	return false

}
