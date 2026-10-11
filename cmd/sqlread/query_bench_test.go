package main

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"

	"github.com/donatj/sqlread"
)

func BenchmarkExecQuery(b *testing.B) {
	input := "CREATE TABLE `records` (`id` int);\nINSERT INTO `records` VALUES " +
		strings.Repeat("(1),", 20000) + "(2);"
	source := strings.NewReader(input)
	l, items := sqlread.Lex(source)
	done := make(chan struct{})
	go func() { defer close(done); l.Run(sqlread.StartState) }()
	defer func() { l.Stop(); <-done }()
	summary := sqlread.NewSummaryParser()
	if err := sqlread.Parse(items).Run(summary.ParseStart); err != nil {
		b.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		limit *sqlread.QueryLimit
	}{{"all_rows", nil}, {"limit_10", &sqlread.QueryLimit{Count: 10}}} {
		b.Run(test.name, func(b *testing.B) {
			qry := sqlread.Query{Table: "records", Columns: []string{"*"}, Limit: test.limit}
			writer := csv.NewWriter(io.Discard)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := execQuery(summary.Tree, qry, source, writer); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
