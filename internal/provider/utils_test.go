// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"testing"
)

func TestSnakeToCamel(t *testing.T) {
	input := make([]string, 3)
	input[0] = "i_am_snake_case"
	input[1] = "IAmUpperCamelCase"
	input[2] = "iAmAlreadyCamelCase"

	expected := []string{"iAmSnakeCase", "iAmUpperCamelCase", "iAmAlreadyCamelCase"}

	for it, in := range input {
		result := SnakeToCamel(in)

		if result != expected[it] {
			t.Errorf("Failed [%s]: result (%s) did not match expected (%s)", in, result, expected[it])
		}
	}
}

func TestCamelToSnake(t *testing.T) {
	input := make([]string, 3)
	input[0] = "i_am_snake_case"
	input[1] = "IAmUpperCamelCase"
	input[2] = "iAmLowerCamelCase"

	expected := []string{"i_am_snake_case", "i_am_upper_camel_case", "i_am_lower_camel_case"}

	for it, in := range input {
		result := CameltoSnake(in)

		if result != expected[it] {
			t.Errorf("Failed [%s]: result (%s) did not match expected (%s)", in, result, expected[it])
		}
	}
}

func TestIsEmail(t *testing.T) {
	type testCase struct {
		input string
		want  bool
	}

	tests := []testCase{
		{
			input: "",
			want:  false,
		},
		{
			input: "1234567890987654321",
			want:  false,
		},
		{
			input: "example.com",
			want:  false,
		},
		{
			input: "user@example.com",
			want:  true,
		},
	}

	for _, tc := range tests {
		got := isEmail(tc.input)
		if tc.want != got {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
