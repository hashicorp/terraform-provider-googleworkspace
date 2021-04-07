package googleworkspace

import (
	"context"
	"regexp"
	"time"

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
	regexp.MatchString(err.Error(), "timed out while waiting")

	return true
}
