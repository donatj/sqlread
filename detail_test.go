package sqlread

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

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

func TestStringValueControlEscapes(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"backspace", "\\b", "\b"},
		{"newline", "\\n", "\n"},
		{"carriage return", "\\r", "\r"},
		{"tab", "\\t", "\t"},
		{"control Z", "\\Z", "\x1a"},
		{"mixed", "\\b\\n\\r\\t\\Z", "\b\n\r\t\x1a"},
		{"NUL", "\\0", "\x00"},
		{"backslash", "\\\\", "\\"},
		{"single quote", "\\'", "'"},
		{"double quote", "\\\"", "\""},
	}

	for _, quote := range []string{"'", "\""} {
		for _, test := range tests {
			t.Run(quote+test.name, func(t *testing.T) {
				input := quote + test.input + quote
				actual, err := stringValue(LexItem{Val: input})
				if actual != test.output || err != nil {
					t.Errorf("stringValue(%q) = %q, %v; want %q, nil", input, actual, err, test.output)
				}
			})
		}
	}
}

func TestInsertDetailControlEscapes(t *testing.T) {
	for _, quote := range []string{"'", "\""} {
		t.Run(quote, func(t *testing.T) {
			input := "INSERT INTO `records` VALUES (" +
				quote + "left\\nright\\tend" + quote + "," +
				quote + "\\b\\r\\Z" + quote + "),(" +
				quote + "left\\\\nright\\\\tend" + quote + "," +
				quote + "\\\\b\\\\r\\\\Z" + quote + ");"
			want := [][]string{
				{"left\nright\tend", "\b\r\x1a"},
				{"left\\nright\\tend", "\\b\\r\\Z"},
			}

			l, items := Lex(strings.NewReader(input))
			go l.Run(StartState)
			d := NewInsertDetailParser()
			if err := Parse(items).Run(d.ParseStart); err != nil {
				t.Fatal(err)
			}

			var actual [][]string
			for row := range d.Out {
				actual = append(actual, row)
			}
			if !reflect.DeepEqual(actual, want) {
				t.Errorf("parsed rows = %q; want %q", actual, want)
			}
		})
	}
}

func TestInsertDetailParserStop(t *testing.T) {
	input := "INSERT INTO records VALUES " + strings.Repeat("(1),", 200) + "(1);"
	l, items := Lex(strings.NewReader(input))
	go l.Run(StartState)

	detail := NewInsertDetailParser()
	done := make(chan error, 1)
	go func() {
		done <- Parse(items).Run(detail.ParseStart)
	}()

	for range 10 {
		<-detail.Out
	}

	detail.Stop()
	l.Stop()
	for range detail.Out {
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("parser did not stop")
	}
}
