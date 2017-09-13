package sqlread

import "testing"

func TestStringValue(t *testing.T) {
	tests := []struct {
		input  string
		output string
		err    error
	}{
		{`no delimiters`, "", errorInvalidString},
		{`'mismatched delimiters"`, "", errorInvalidString},

		{`'I am made of sauce'`, "I am made of sauce", nil},
		{`"I have lived a happy life free of sauce"`, "I have lived a happy life free of sauce", nil},
		{`"foo""bar"`, `foo"bar`, nil},
		{`"mismatched at end""`, ``, errorInvalidString},

		{`"foo\"bar"`, `foo"bar`, nil},

		{`"foo\\tbar"`, `foo\tbar`, nil},
		{`"foo\xbar"`, ``, errorInvalidString},
	}

	for _, test := range tests {
		actual, err := stringValue(LexItem{Val: test.input})

		if actual != test.output || err != test.err {
			t.Errorf(`stringValue(LexItem{Val: %#v}) = %#v, %#v; want %#v, %#v`, test.input, actual, err, test.output, test.err)
		}
	}
}
