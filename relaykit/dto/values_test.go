package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntValueUnmarshalNumericForms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IntValue
	}{
		{name: "integer", input: `5`, expected: 5},
		{name: "decimal integer", input: `5.0`, expected: 5},
		{name: "fractional number", input: `13.93`, expected: 13},
		{name: "negative fractional number", input: `-2.7`, expected: -2},
		{name: "integer string", input: `"5"`, expected: 5},
		{name: "decimal integer string", input: `"5.0"`, expected: 5},
		{name: "fractional string", input: `"13.93"`, expected: 13},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual IntValue
			err := json.Unmarshal([]byte(test.input), &actual)

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestIntValueUnmarshalRejectsInvalidValues(t *testing.T) {
	tests := []string{
		`"not-a-number"`,
		`"NaN"`,
		`1e100`,
		`true`,
		`{}`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var value IntValue
			err := json.Unmarshal([]byte(input), &value)

			require.Error(t, err)
		})
	}
}
