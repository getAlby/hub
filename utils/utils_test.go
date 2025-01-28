package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommandLine(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name            string
		input           string
		expectedSuccess []string
		expectedError   string
	}

	// When called by the API, the first argument of the command input is actually the command name
	testCases := []testCase{
		{
			name:            "empty input",
			input:           "",
			expectedSuccess: []string{},
			expectedError:   "",
		},
		{
			name:            "single argument",
			input:           "arg1",
			expectedSuccess: []string{"arg1"},
			expectedError:   "",
		},
		{
			name:            "multiple arguments",
			input:           "arg1 arg2 arg3",
			expectedSuccess: []string{"arg1", "arg2", "arg3"},
			expectedError:   "",
		},
		{
			name:            "multiple arguments with extra whitespace",
			input:           "  arg1\targ2   arg3  ",
			expectedSuccess: []string{"arg1", "arg2", "arg3"},
			expectedError:   "",
		},
		{
			name:            "multiple arguments with quotes and escaping",
			input:           `"arg 1" arg2 "arg\"3"`,
			expectedSuccess: []string{"arg 1", "arg2", `arg"3`},
			expectedError:   "",
		},
		{
			name:            "unquoted escaped whitespace",
			input:           `arg\ 1 arg2`,
			expectedSuccess: []string{"arg 1", "arg2"},
			expectedError:   "",
		},
		{
			name:            "escaped JSON",
			input:           `{\"hello\":\"world\"}`,
			expectedSuccess: []string{`{"hello":"world"}`},
			expectedError:   "",
		},
		{
			name:            "escaped JSON with space",
			input:           `"{\"hello\": \"world\"}"`,
			expectedSuccess: []string{`{"hello": "world"}`},
			expectedError:   "",
		},
		{
			name:            "unclosed quote",
			input:           `"arg 1", "arg2", "arg\"3`,
			expectedSuccess: nil,
			expectedError:   "unexpected end of string",
		},
		{
			name:            "three quotes",
			input:           `"""`,
			expectedSuccess: nil,
			expectedError:   "unexpected end of string",
		},
		{
			name:            "unfinished escape",
			input:           `arg\`,
			expectedSuccess: nil,
			expectedError:   "unexpected end of string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parsedArgs, err := ParseCommandLine(tc.input)
			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSuccess, parsedArgs)
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Empty(t, parsedArgs)
			}
		})
	}
}
