package sqlread

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestQueryParserLimit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  QueryLimit
	}{
		{
			name:  "count",
			input: "SELECT * FROM records LIMIT 2;",
			want:  QueryLimit{Count: 2},
		},
		{
			name:  "offset and count",
			input: "select name from records limit 1, 2;",
			want:  QueryLimit{Offset: 1, Count: 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := parseUserQuery(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if len(query.Queries) != 1 {
				t.Fatalf("parsed queries = %d; want 1", len(query.Queries))
			}
			if !reflect.DeepEqual(query.Queries[0].Limit, &test.want) {
				t.Errorf("limit = %#v; want %#v", query.Queries[0].Limit, &test.want)
			}
		})
	}
}

func TestQueryParserLimitRejectsNonIntegers(t *testing.T) {
	for _, input := range []string{
		"SELECT * FROM records LIMIT -1;",
		"SELECT * FROM records LIMIT 1.5;",
		"SELECT * FROM records LIMIT 1, count;",
	} {
		t.Run(input, func(t *testing.T) {
			_, err := parseUserQuery(input)
			if err == nil {
				t.Fatal("parseUserQuery() returned nil error")
			}
		})
	}
}

func TestQueryParserDoesNotReadPastSemicolon(t *testing.T) {
	for _, query := range []string{
		"SELECT * FROM records;",
		"SELECT * FROM records INTO OUTFILE 'records.csv';",
	} {
		t.Run(query, func(t *testing.T) {
			input := &readPastEndReader{input: query}

			_, err := parseUserQueryReader(input)
			if err != nil {
				t.Fatal(err)
			}

			if input.readPastEnd {
				t.Fatal("parser read past the statement terminator")
			}
		})
	}
}

func TestQueryLimitIncludes(t *testing.T) {
	tests := []struct {
		name  string
		limit QueryLimit
		row   uint64
		want  bool
	}{
		{name: "before offset", limit: QueryLimit{Offset: 1, Count: 2}, row: 0, want: false},
		{name: "at offset", limit: QueryLimit{Offset: 1, Count: 2}, row: 1, want: true},
		{name: "within count", limit: QueryLimit{Offset: 1, Count: 2}, row: 2, want: true},
		{name: "after count", limit: QueryLimit{Offset: 1, Count: 2}, row: 3, want: false},
		{name: "zero count", limit: QueryLimit{Count: 0}, row: 0, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := test.limit.Includes(test.row); actual != test.want {
				t.Errorf("Includes(%d) = %t; want %t", test.row, actual, test.want)
			}
		})
	}
}

func parseUserQuery(input string) (QueryTree, error) {
	return parseUserQueryReader(strings.NewReader(input))
}

func parseUserQueryReader(input io.ReaderAt) (QueryTree, error) {
	l, items := Lex(input)
	intp := &Intp{}
	go l.Run(intp.StartIntpState)

	query := NewQueryParser()
	err := Parse(items).Run(query.ParseStart)
	return query.Tree, err
}

type readPastEndReader struct {
	input       string
	readPastEnd bool
}

func (r *readPastEndReader) ReadAt(b []byte, off int64) (int, error) {
	if off+int64(len(b)) > int64(len(r.input)) {
		r.readPastEnd = true
	}

	if off >= int64(len(r.input)) {
		return 0, io.EOF
	}

	n := copy(b, r.input[off:])
	if n != len(b) {
		return n, io.EOF
	}

	return n, nil
}
