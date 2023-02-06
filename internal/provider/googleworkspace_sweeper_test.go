// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// List of prefixes used for test resource names
var testResourcePrefixes = []string{
	"tf-test",
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func isSweepableTestResource(resourceName string) bool {
	for _, p := range testResourcePrefixes {
		if strings.HasPrefix(resourceName, p) {
			return true
		}
	}
	return false
}
