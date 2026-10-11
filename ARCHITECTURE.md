# Architecture

sqlread queries MySQL dump files without importing them into a database. Its
primary design goal is to process very large dumps on machines with limited
memory. The dump may be many times larger than available memory. Low memory use
is a core requirement that every feature must preserve.

The existing tool has already been used successfully under these conditions.
Future changes must preserve that capability. Memory requirements must not grow
in proportion to the size of the dump.

**Treat the file as memory.** Scan it to learn where data lives. Keep compact
references to that data. Read the referenced bytes when they are needed.

This document records that foundation and defines requirements for future work.
The final section identifies areas to measure or refine within the existing
implementation.

## Foundation: Lexical Scanning in Go

The scanner design must remain firmly based on Rob Pike's
[Lexical Scanning in Go](https://go.dev/talks/2011/lex.slide#1), presented on
30 August 2011. State functions are the foundation:

```go
type state func(*lexer) state
```

Each function handles one scanning state. It returns the next state. Returning
`nil` ends the scan. Small helpers handle lookahead, character consumption,
backup, and token emission. See slides 18–24 and 30–36.

The talk first connects the lexer and its consumer through a channel. Slides
39–43 then hide that channel behind a next-token method. That version keeps the
state functions without requiring a separate goroutine. Both interfaces fit
this architecture. The design does not require parallel execution.

Pike's example scans an input string. sqlread must adapt that design to a file
with bounded read buffers. Keeping the whole input string is unsuitable for
large dumps. File references and the memory rules below are sqlread-specific
requirements.

## The file is the backing store

The dump remains the authoritative copy of its contents. The index describes
locations in that file. It must not become a second copy of the rows.

- Represent source locations with byte offsets and lengths. Define their bounds
  consistently. Translate section-relative positions before storing them in a
  file-wide index.
- Keep the open source file available while its references are in use.
- Use `io.ReaderAt` and bounded sections to revisit data. Read nearby bytes
  through a small reusable buffer.
- Retain names and schema information only where needed. Keep large definitions,
  comments, literals, and data blocks as file spans.
- Prefer reading bytes again over retaining an unbounded collection of values.

Treating the file as memory does not require memory mapping. It describes where
the data lives. Any access strategy must still meet the memory budget.

The source must remain unchanged during a session. Detect changes where
practical. Fail clearly if the index is no longer valid. File size alone cannot
establish that the contents are unchanged.

## Responsibilities

These are logical boundaries. They do not require a separate package for every
row in this table.

| Component | Responsibility |
| --- | --- |
| Source reader | Read bounded regions. Maintain file positions. Report read errors. |
| Dump scanner | Recognize dump tokens through state functions. Preserve source spans. |
| Summary parser | Build the table catalog and references to data regions. |
| Row reader | Decode requested values from a referenced region. Release data as it is consumed. |
| User-query parser | Turn a user statement into a query description. Validate supported syntax. |
| Query executor | Select source regions and columns. Apply supported operations. Stream results. |
| Cache | Persist derived index data. Validate it against the source. |
| Command | Own flags, terminal input, output files, and user-facing errors. |

The scanner recognizes tokens. Parsers own their valid combinations. Dump
syntax and user-query syntax have different purposes. They may share low-level
scanning helpers without sharing all grammar rules.

Execution must be callable without a terminal. Core code returns errors to its
caller. It must not exit the process or depend on global command flags.

## State and simplicity

The lexer and parser need state to track position and progress. Keep that state
explicit and local to the component that owns it.

Elsewhere, prefer functions and methods with explicit inputs and results. Avoid
hidden dependencies, mutable package variables, and fields that retain data
needed only during a single call.

Fully stateless code is not a requirement. Use local mutation and stateful types
when they make Go code clearer or support efficient streaming. Do not add layers,
repeated copying, or awkward parameter passing solely to eliminate state. Keep
necessary state small. Give it a clear owner and a limited lifetime.

## Initial scan

Scan the dump in file order. Discover table definitions and data regions.
Record enough information to return to those regions later.

The scan must recognize quoting, escaping, and statement boundaries. It cannot
find boundaries by blindly searching for semicolons. It does not need to retain
decoded row values or build a syntax tree for the entire dump.

Index memory also needs a bound. A compact entry for every `INSERT` can still
consume substantial memory when a dump contains millions of small statements.
Use coarse ranges where appropriate. Store larger indexes on disk with bounded
access. Do not assume that metadata is always small.

## User input and execution

The terminal layer identifies a complete user statement before passing it to
the query parser. Statement framing must respect quoted strings and identifiers.
Parser lookahead must not request another line after the terminating semicolon.
Input accumulation and history need explicit size limits.

The executor resolves a query against the table catalog. It visits the relevant
file regions in order. It reads only as far as the operation requires. It writes
results as they become available.

For `LIMIT offset, count`, count rows across all data regions for the table.
Discard rows before the offset. Emit at most the requested count. Stop the active
scan once that count is reached. A zero count requires no row decoding.

Read and decode only the values required by an operation where practical.
Preserve the distinction between `NULL` and string contents until the output
format handles them. Do not accumulate a result set before writing it.

## Memory rules

Working memory must stay bounded as dump size and result count grow. Schema and
index growth must fit the same budget or move to bounded storage on disk.

Every buffer, queue, and retained collection must have an understood bound.
Document the reason for its capacity. Measure its cost before increasing it.

- Keep token queues small. When a consumer slows down, the producer must wait.
  Larger queues need a measured throughput benefit within the memory budget.
- Bound queued bytes as well as item counts. One token can contain a huge value.
- Emit source spans when text is not immediately needed. Avoid allocating a
  string for every literal during indexing.
- Process large values in chunks or through file references. Holding one row
  at a time is not enough if that row can exceed available memory.
- Reuse bounded scratch space. Do not let a small retained slice keep a large
  backing allocation alive. State the lifetime of any borrowed buffer.
- Stream output through a bounded writer. Check write and flush errors.
- Apply the same limits while loading or rebuilding a cache. A cache hit must
  not bypass the memory budget.

If an operation cannot yet handle a value within its budget, return a clear
size-limit error. Do not silently allocate without a bound or truncate data.

Future features must preserve these rules. Filtering can inspect rows as they
arrive. Sorting may require temporary files and bounded merge buffers. A feature
that needs all rows in memory requires a different implementation before it can
be supported.

## Completion and cancellation

Each scan has one owner. That owner must be able to stop it after a limit,
user cancellation, read failure, parse failure, or output failure.

If channels and goroutines are used, keep them bounded. Cancellation must reach
blocked senders and receivers. Long scanning loops must also observe it. The
owner waits for its workers to finish before reusing shared buffers or closing
the source.

A synchronous next-token interface is also valid. It must stop requesting data
when its consumer stops. Choose concurrency for a measured benefit or clearer
control flow. Keep the state-function design in either case.

End of file, malformed input, and cancellation must have distinct outcomes.
Every path must terminate. Report useful source positions without copying huge
tokens into error messages.

## Cache contract

The cache is disposable derived data. Rebuilding it from the dump must always
remain possible. Cache format versions and source validation must be explicit.
File size and modification metadata can detect common changes. They do not
prove content identity. Stronger validation must scan bytes in bounded chunks.

Store a cache through a temporary file before replacing the completed cache.
Missing, stale, or corrupt cache data should trigger a rebuild. Disabling the
cache must bypass both loading and storing it.

## Validation

Test the boundaries between components with small dump fixtures. Cover multiple
data regions, quoted delimiters, partial input, cancellation, and output failures.
Check the resulting rows and completion behavior through the real executor.

Measure peak live memory during indexing, cache loading, and query execution.
Increase row counts and dump size while holding the schema fixed. Then vary
statement count, schema size, and individual value size separately. This exposes
growth in both retained data and metadata.

Track application allocations separately from operating-system file caching.
Measure throughput alongside memory. Any larger buffer needs evidence that its
benefit is worth its retained memory.

## Areas to refine

The current code already uses state functions and file positions to support
large dumps with limited memory. These areas merit measurement or refinement
as the tool evolves:

- [scanner.go](scanner.go) allocates channels for 2,000,000 tokens. Token emission
  copies source text into strings.
- [cmd/sqlread/main.go](cmd/sqlread/main.go) configures a 100,000,000-byte reader
  buffer. Its capacity needs a measured justification against the memory budget.
- [summary.go](summary.go) retains all data-region references in memory.
  [mapcache/cache.go](mapcache/cache.go) decodes the entire cached summary.
- [detail.go](detail.go) builds complete row values as strings. Large individual
  values still need a bounded representation.
- [cmd/sqlread/readlinewrap.go](cmd/sqlread/readlinewrap.go) connects parser reads
  directly to terminal input.
- [cmd/sqlread/query.go](cmd/sqlread/query.go) keeps execution in the command
  package. Worker completion and output errors need a consistent owner.

Refinements must preserve the tool's ability to process very large dumps with
limited memory. Keep the state-function foundation and the file as the backing
store.
