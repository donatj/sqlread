package sqlread

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestExecuteQuery(t *testing.T) {
	input := "CREATE TABLE `records` (\n" +
		"  `id` int NOT NULL,\n" +
		"  `name` varchar(255) NOT NULL\n" +
		");\n" +
		"INSERT INTO `records` VALUES (1,'one'),(2,'two');\n" +
		"INSERT INTO `records` VALUES (3,'three'),(4,'four');"
	source := strings.NewReader(input)
	tree := indexExecuteTestDump(t, source)

	var rows [][]string
	err := ExecuteQuery(tree, Query{
		Table:   "records",
		Columns: []string{"name", "id"},
		Limit:   &QueryLimit{Offset: 1, Count: 2},
	}, source, func(row []string) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"two", "2"}, {"three", "3"}}; !reflect.DeepEqual(rows, want) {
		t.Errorf("rows = %#v; want %#v", rows, want)
	}
}

func TestExecuteQueryStopsOnOutputError(t *testing.T) {
	input := "CREATE TABLE `records` (`id` int);\nINSERT INTO `records` VALUES (1),(2);"
	source := strings.NewReader(input)
	tree := indexExecuteTestDump(t, source)
	want := errors.New("output failed")
	calls := 0

	err := ExecuteQuery(tree, Query{Table: "records", Columns: []string{"*"}}, source, func([]string) error {
		calls++
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v; want %v", err, want)
	}
	if calls != 1 {
		t.Errorf("write calls = %d; want 1", calls)
	}
}

func TestExecuteQueryZeroLimitReadsNoRows(t *testing.T) {
	input := "CREATE TABLE `records` (`id` int);\nINSERT INTO `records` VALUES (1),(2);"
	source := strings.NewReader(input)
	tree := indexExecuteTestDump(t, source)
	calls := 0

	err := ExecuteQuery(tree, Query{
		Table:   "records",
		Columns: []string{"*"},
		Limit:   &QueryLimit{},
	}, source, func([]string) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Errorf("write calls = %d; want 0", calls)
	}
}

func indexExecuteTestDump(t *testing.T, source *strings.Reader) SummaryTree {
	t.Helper()
	l, items := Lex(source)
	done := make(chan struct{})
	go func() {
		defer close(done)
		l.Run(StartState)
	}()
	defer func() {
		l.Stop()
		<-done
	}()

	summary := NewSummaryParser()
	if err := Parse(items).Run(summary.ParseStart); err != nil {
		t.Fatal(err)
	}
	return summary.Tree
}
