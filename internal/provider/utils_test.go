package provider

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
