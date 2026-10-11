package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/donatj/sqlread"
	"github.com/donatj/sqlread/mapcache"
)

func TestQueryWorkflow(t *testing.T) {
	file := queryFixture(t)
	tree := indexDump(t, file)
	cache := mapcache.New(file)
	if err := cache.Store(tree); err != nil {
		t.Fatal(err)
	}
	cached, err := cache.Get()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		query string
		want  string
	}{
		{"SELECT id FROM records;", "1\n2\n3\n4\n5\n6\n"},
		{"SELECT * FROM records LIMIT 1;", "1,alpha\n"},
		{"SELECT * FROM records LIMIT 3;", "1,alpha\n2,\"two,parts\"\n3,\"three\nlines\"\n"},
		{"select id from records limit 1, 2;", "2\n3\n"},
		{"SELECT id FROM records LIMIT 2, 2;", "3\n4\n"},
		{"SELECT name, id FROM records LIMIT 3, 2;", "four,4\nfive,5\n"},
		{"SELECT id FROM records LIMIT 5, 10;", "6\n"},
		{"SELECT id FROM records LIMIT 6, 1;", ""},
		{"SELECT id FROM records LIMIT 0;", ""},
		{"SELECT id FROM records LIMIT 5, 0;", ""},
		{"SELECT id FROM records LIMIT 18446744073709551615, 2;", ""},
		{"SELECT id FROM records LIMIT 1, 18446744073709551615;", "2\n3\n4\n5\n6\n"},
		{"SELECT id FROM other LIMIT 2;", "99\n"},
	}
	for _, indexed := range []struct {
		name string
		tree sqlread.SummaryTree
	}{{"fresh", tree}, {"cached", cached}} {
		t.Run(indexed.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.query, func(t *testing.T) {
					qry := parseQuery(t, test.query)
					var output bytes.Buffer
					if err := execQuery(indexed.tree, qry, file, csv.NewWriter(&output)); err != nil {
						t.Fatal(err)
					}
					if got := output.String(); got != test.want {
						t.Errorf("output = %q; want %q", got, test.want)
					}
				})
			}
		})
	}
}

func TestQueryStopsReading(t *testing.T) {
	// Many rows in one INSERT ensure a small LIMIT must stop inside a data block.
	input := "CREATE TABLE `records` (`id` int);\nINSERT INTO `records` VALUES " +
		strings.Repeat("(1),", 10000) + "(2);"
	tree := indexDump(t, strings.NewReader(input))
	for _, query := range []string{
		"SELECT * FROM records LIMIT 0;",
		"SELECT * FROM records LIMIT 500, 0;",
		"SELECT * FROM records LIMIT 1;",
	} {
		t.Run(query, func(t *testing.T) {
			source := &queryReader{ReaderAt: strings.NewReader(input)}
			qry := parseQuery(t, query)
			var output bytes.Buffer
			if err := execQuery(tree, qry, source, csv.NewWriter(&output)); err != nil {
				t.Fatal(err)
			}
			// This also detects a lexer worker still reading after execution returns.
			reads := source.reads.Load()
			if qry.Limit.Count == 0 {
				if reads != 0 || output.Len() != 0 {
					t.Fatalf("zero limit: reads = %d, output = %q", reads, output.String())
				}
			} else if output.String() != "1\n" || source.farthest.Load() >= int64(len(input)/2) {
				t.Fatalf("limit did not stop early: offset = %d, output = %q", source.farthest.Load(), output.String())
			}
			source.closed.Store(true)
			if reads != source.reads.Load() {
				t.Fatal("query returned before its reads completed")
			}
		})
	}
}

func TestQueryFailures(t *testing.T) {
	file := queryFixture(t)
	tree := indexDump(t, file)
	for _, query := range []string{
		"SELECT missing FROM records LIMIT 1;",
		"SELECT * FROM missing LIMIT 1;",
	} {
		t.Run(query, func(t *testing.T) {
			if err := execQuery(tree, parseQuery(t, query), file, csv.NewWriter(io.Discard)); err == nil {
				t.Fatal("expected a query error")
			}
		})
	}

	t.Run("source read", func(t *testing.T) {
		want := errors.New("source read failed")
		source := &queryReader{ReaderAt: file, err: want}
		err := execQuery(tree, parseQuery(t, "SELECT * FROM records LIMIT 1;"), source, csv.NewWriter(io.Discard))
		if !errors.Is(err, want) {
			t.Fatalf("error = %v; want %v", err, want)
		}
	})

	t.Run("flush", func(t *testing.T) {
		want := errors.New("output failed")
		writer := csv.NewWriter(errorWriter{want})
		err := execQuery(tree, parseQuery(t, "SELECT * FROM records LIMIT 1;"), file, writer)
		if !errors.Is(err, want) {
			t.Fatalf("error = %v; want %v", err, want)
		}
	})

	t.Run("write", func(t *testing.T) {
		want := errors.New("row write failed")
		source := &queryReader{ReaderAt: file}
		writes := 0
		err := sqlread.ExecuteQuery(tree, parseQuery(t, "SELECT * FROM records;"), source, func([]string) error {
			writes++
			return want
		})
		source.closed.Store(true)
		if !errors.Is(err, want) || writes != 1 {
			t.Fatalf("error = %v, writes = %d; want %v and 1 write", err, writes, want)
		}
	})

	for _, test := range []struct {
		row  string
		want string
	}{
		{"(1)", "row has 1 columns; expected 2"},
		{"(1,'bad\\x')", "invalid string"},
		{"(1,unsupported)", "TIllegal"},
		{"(1,'unterminated", "EOF"},
	} {
		t.Run(test.row, func(t *testing.T) {
			// Keep valid table metadata. Point its data span at the entire malformed INSERT.
			input := "CREATE TABLE `records` (`id` int, `name` text);\n"
			valid := input + "INSERT INTO `records` VALUES (1,'valid');"
			indexed := indexDump(t, strings.NewReader(valid))
			bad := input + "INSERT INTO `records` VALUES " + test.row + ";"
			indexed["records"].DataLocs[0].End.Pos = int64(len(bad) - 1)
			err := execQuery(indexed, parseQuery(t, "SELECT * FROM records;"), strings.NewReader(bad), csv.NewWriter(io.Discard))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v; want %q", err, test.want)
			}
		})
	}
}

func queryFixture(t testing.TB) *os.File {
	t.Helper()
	data, err := os.ReadFile("../../testfixtures/queries.sql")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "dump.sql")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { file.Close() })
	return file
}

func indexDump(t testing.TB, source io.ReaderAt) sqlread.SummaryTree {
	t.Helper()
	l, items := sqlread.Lex(source)
	done := make(chan struct{})
	go func() { defer close(done); l.Run(sqlread.StartState) }()
	defer func() { l.Stop(); <-done }()
	summary := sqlread.NewSummaryParser()
	if err := sqlread.Parse(items).Run(summary.ParseStart); err != nil {
		t.Fatal(err)
	}
	return summary.Tree
}

func parseQuery(t testing.TB, input string) sqlread.Query {
	t.Helper()
	l, items := sqlread.Lex(strings.NewReader(input))
	done := make(chan struct{})
	intp := &sqlread.Intp{}
	go func() { defer close(done); l.Run(intp.StartIntpState) }()
	defer func() { l.Stop(); <-done }()
	parser := sqlread.NewQueryParser()
	if err := sqlread.Parse(items).Run(parser.ParseStart); err != nil {
		t.Fatal(err)
	}
	if len(parser.Tree.Queries) != 1 {
		t.Fatalf("parsed %d queries; want 1", len(parser.Tree.Queries))
	}
	return parser.Tree.Queries[0]
}

type queryReader struct {
	io.ReaderAt
	reads    atomic.Int64
	farthest atomic.Int64
	closed   atomic.Bool
	err      error
}

func (r *queryReader) ReadAt(b []byte, off int64) (int, error) {
	if r.closed.Load() {
		panic("source used after query returned")
	}
	r.reads.Add(1)
	if off+int64(len(b)) > r.farthest.Load() {
		r.farthest.Store(off + int64(len(b)))
	}
	if r.err != nil {
		return 0, r.err
	}
	return r.ReaderAt.ReadAt(b, off)
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }
