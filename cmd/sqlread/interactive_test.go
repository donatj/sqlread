package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestInteractiveQueries(t *testing.T) {
	file := queryFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInteractiveProcess$", file.Name())
	cmd.Env = append(os.Environ(), "SQLREAD_TEST_INTERACTIVE=1")
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		stdin.Close()
		cancel()
		cmd.Wait()
	}()

	reader := bufio.NewReader(stdout)
	for _, test := range []struct {
		query string
		rows  []string
	}{
		{"SELECT id FROM other;", []string{"99"}},
		{"SELECT id FROM records LIMIT 1, 2;", []string{"2", "3"}},
		{"SELECT id FROM records LIMIT 0;\nSELECT id FROM other;", []string{"99"}},
	} {
		if _, err := fmt.Fprintln(stdin, test.query); err != nil {
			t.Fatal(err)
		}
		// Keep stdin open. Each query must complete without another terminator or EOF.
		for _, want := range test.rows {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("query %q did not produce %q: %v", test.query, want, err)
			}
			if strings.TrimSpace(line) != want {
				t.Fatalf("query %q: output = %q; want %q", test.query, line, want)
			}
		}
	}
	if _, err := fmt.Fprintln(stdin, "QUIT;"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("interactive process: %v\n%s", err, diagnostics.String())
	}
}

func TestInteractiveProcess(t *testing.T) {
	if os.Getenv("SQLREAD_TEST_INTERACTIVE") != "1" {
		return
	}
	main()
	os.Exit(0)
}
