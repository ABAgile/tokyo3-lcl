# lcl - the "water of life" for mental synchronization

[![Release](https://img.shields.io/github/v/release/abagile/tokyo3-lcl?sort=semver&logo=Go&color=%23007D9C)](https://github.com/abagile/tokyo3-lcl/releases)
[![Test](https://github.com/abagile/tokyo3-lcl/actions/workflows/test.yml/badge.svg)](https://github.com/abagile/tokyo3-lcl/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/abagile/tokyo3-lcl.svg)](https://pkg.go.dev/github.com/abagile/tokyo3-lcl)
[![Go Report Card](https://goreportcard.com/badge/github.com/abagile/tokyo3-lcl)](https://goreportcard.com/report/github.com/abagile/tokyo3-lcl)
[![codecov](https://codecov.io/gh/abagile/tokyo3-lcl/branch/main/graph/badge.svg)](https://codecov.io/gh/abagile/tokyo3-lcl)

A generic Go utility library providing small, handy building blocks for
application code.

**Requires Go 1.26+**

```sh
go get github.com/abagile/tokyo3-lcl
```

---

## Design philosophy

### Simple slice-in/slice-out helpers

The collection helpers in `slices.go` operate directly on concrete slices using
plain Go loops. This keeps the package dependency-light and makes allocation and
control flow easy to read.

Filtering, grouping, partitioning, and padding functions preserve named slice
types through `S ~[]T`. Mapping functions return `[]R` because the element type
can change.

These helpers are intentionally eager: each function consumes its input and
returns a concrete slice or value. Chaining helpers is therefore easy to read,
but it allocates intermediate slices:

```go
activeNames := Map(
    Filter(users, func(u User) bool { return u.Active }),
    func(u User) string { return u.Name },
)
```

If you need lazy iterator pipelines, non-slice sources, or allocation-free
composition across multiple transformations, use Go's `iter` package directly
or a library such as
[`github.com/BooleanCat/go-functional`](https://github.com/BooleanCat/go-functional).

### Small panic-on-failure helpers, not a Result monad

`result.go` keeps normal Go `(T, error)` values visible and provides small
`Must*` helpers for CLI tools and startup paths where a failure should abort
immediately. For normal request/runtime code, prefer explicit error handling.

---

## slices.go — Functional slice operations

### Function naming conventions

Each operation comes in up to three variants:

| Suffix | Meaning |
|---|---|
| *(none)* | Plain value callback |
| `2` | Index-aware callback receiving `(index, value)` |
| `Error` | Callback may return an `error`; iteration stops on the first non-nil error |

Reach for the plain variant by default; add `2` when you need the position, and
add `Error` when the callback can fail.

### Filtering

| Function | Description |
|---|---|
| `Filter(xs, pred)` | Keep elements where `pred` is true |
| `Filter2(xs, pred)` | Like `Filter` but predicate receives `(index, value)` |
| `FilterError(xs, pred)` | Like `Filter` but predicate may return an error |
| `Exclude(xs, pred)` | Keep elements where `pred` is false |
| `Exclude2(xs, pred)` | Like `Exclude` with index-aware predicate |
| `ExcludeError(xs, pred)` | Like `Exclude` but predicate may return an error |

```go
evens := Filter([]int{1, 2, 3, 4}, func(n int) bool {
    return n%2 == 0
})
// []int{2, 4}

oddPositions := Exclude2([]string{"a", "b", "c", "d"}, func(i int, _ string) bool {
    return i%2 == 0
})
// []string{"b", "d"}
```

### Transformation

| Function | Description |
|---|---|
| `Map(xs, mapper)` | Transform each element |
| `Map2(xs, mapper)` | Transform with `(index, value)` mapper |
| `MapError(xs, mapper)` | Transform; stops on first error |

```go
labels := Map([]int{10, 20, 30}, func(n int) string {
    return fmt.Sprintf("value=%d", n)
})

indexed := Map2([]string{"a", "b"}, func(i int, s string) string {
    return fmt.Sprintf("%d:%s", i, s)
})
```

### Folding

| Function | Description |
|---|---|
| `Fold(xs, accum, initial)` | Left fold |
| `Fold2(xs, accum, initial)` | Left fold with index |
| `FoldError(xs, accum, initial)` | Left fold; stops on first error |
| `FoldRight(xs, accum, initial)` | Right fold |
| `FoldRight2(xs, accum, initial)` | Right fold with index |
| `FoldRightError(xs, accum, initial)` | Right fold; stops on first error |

```go
sum := Fold([]int{1, 2, 3}, func(acc, n int) int {
    return acc + n
}, 0)
// 6

reversed := FoldRight([]string{"a", "b", "c"}, func(acc, s string) string {
    return acc + s
}, "")
// "cba"
```

### Iteration

| Function | Description |
|---|---|
| `ForEach(xs, f)` | Call `f` for every element |
| `ForEach2(xs, f)` | Call `f(index, value)` for every element |
| `ForEachWhile(xs, pred)` | Iterate while `pred` returns true |

```go
ForEach(files, func(path string) {
    log.Println("processing", path)
})

ForEachWhile(events, func(e Event) bool {
    return handle(e) // return false to stop early
})
```

### Grouping

| Function | Description |
|---|---|
| `GroupBy(xs, mapper)` | `map[R]S` — groups share the same key; order within each group is preserved |
| `PartitionBy(xs, mapper)` | `[]S` — like `GroupBy`, but groups are returned in first-seen key order |
| `FrequenciesBy(xs, mapper)` | `map[R]int` — count occurrences by key |

Pass `Id` as the mapper when the element itself is the key:

```go
// group/count by the value itself
GroupBy(words, Id)        // map[string][]string
FrequenciesBy(words, Id)  // map[string]int — equivalent to a word-count
PartitionBy(runs, Id)     // ordered groups, e.g. [1,1,2,2,1] → [[1,1,1],[2,2]]
```

### Padding

| Function | Description |
|---|---|
| `Pad(xs, n)` | Extend slice to length `n` with zero values; panics if `n < 0` |
| `PadWith(xs, n, padding)` | Like `Pad` but fills with `padding` |

```go
Pad([]int{1, 2}, 5)          // []int{1, 2, 0, 0, 0}
PadWith([]int{1, 2}, 5, 99)  // []int{1, 2, 99, 99, 99}
```

---

## result.go — Panic-on-failure helpers

Designed for CLI tools and startup paths where certain failures have no sensible
recovery action — a required environment variable is missing, a database
connection must succeed, or a hard dependency cannot be reached. The helpers
panic with a clear message on failure, or return the value and move on.

```go
func Must[T any](value T, err error) T
func MustMsg[T any](value T, err error, msg string, args ...any) T
func MustPass(err error, msg string, args ...any)
func MustPresent[T comparable](value T, msg string, args ...any) T
```

- **`Must`** — unwraps a `(T, error)` pair. Panics with the original error when
  `err != nil`.
- **`MustMsg`** — like `Must`, but panics with `"<formatted msg>: <err>"`.
- **`MustPass`** — for error-only calls.
- **`MustPresent`** — panics when the value is the zero value; otherwise returns
  it. The value must be `comparable` so the check can use direct `==`.

Examples:

```go
db := Must(sql.Open("pgx", dsn))
MustPass(db.PingContext(ctx), "database ping failed")

redisHost := MustPresent(os.Getenv("REDIS_HOST"), "REDIS_HOST must not be empty")
```

Use `MustMsg` when you already have separate `value, err` variables and want
extra context:

```go
file, err := os.Open(path)
file = MustMsg(file, err, "open %q", path)
```

For request handlers, library APIs, and recoverable failures, prefer normal Go
error returns instead of panics.

---

## math.go — Statistical helpers

### `MinMax`

```go
func MinMax[T cmp.Ordered](a, b T) (min, max T)
```

Returns `(min, max)` regardless of argument order.

### `Clamp`

```go
func Clamp[T cmp.Ordered](value, min, max T) T
```

Clamps `value` into `[min, max]`. Caller must ensure `min <= max`.

### `Mean`

```go
func Mean[T Number](xs []T) (T, bool)
```

Returns the arithmetic mean of the slice. Returns `(zero, false)` for an empty
slice.

### `Mode`

```go
func Mode[T comparable](xs []T) ([]T, bool)
```

Returns all values that appear with the highest frequency. Returns `([], false)`
for an empty slice. Order of results is non-deterministic when multiple values
tie.

---

## types.go — Generic type utilities

### Zero / emptiness

```go
func Empty[T any]() T                          // returns zero value of T
func IsEmpty[T comparable](v T) bool           // v == zero
func IsNotEmpty[T comparable](v T) bool        // v != zero
```

### Coalesce

```go
func Coalesce[T comparable](values ...T) (T, bool)   // first non-zero value
func Coalesced[T comparable](values ...T) T           // first non-zero, or zero
```

### Pointer helpers

```go
func ToPtr[T comparable](v T) *T             // nil if v is zero
func FromPtr[T any](ptr *T, fallback ...T) T // dereference or fallback/zero
```

### Slice conversion

```go
func ToAnySlice[T any](in []T) []any
func FromAnySlice[T any](in []any) ([]T, bool)   // false if any element cannot be asserted to T
```

### Nested data access

`GetIn` and `SetIn` navigate nested data structures using a dot-separated path
string. Slice elements are accessed by numeric index.

```go
func GetIn(data any, path string) (any, error)
func SetIn(data any, path string, value any) error
```

`GetIn` supports `map[string]any`, `map[any]any`, and `[]any` while traversing.
`SetIn` supports `map[string]any` and `[]any`; it automatically creates
intermediate `map[string]any` nodes for missing map keys. It does **not** grow
slices — the index must already be in bounds.

```go
data := map[string]any{
    "user": map[string]any{
        "scores": []any{10, 20, 30},
    },
}

v, _ := GetIn(data, "user.scores.1")  // 20
SetIn(data, "user.scores.1", 99)      // scores[1] = 99
```

### Identity

```go
func Id[T any](v T) T
```

Returns its argument unchanged. Primarily useful as a first-class mapper when
the element itself should be the key — avoiding the need to write
`func(x T) T { return x }` at every call site:

```go
FrequenciesBy(tags, Id)    // count each tag
GroupBy(events, Id)        // bucket identical events together
PartitionBy(tokens, Id)    // ordered groups of equal tokens
```

---

## tuples.go — Tuple types and zip/unzip

### Types

```go
type Tuple2[A, B any]       struct{ A A; B B }
type Tuple3[A, B, C any]    struct{ A A; B B; C C }
type Tuple4[A, B, C, D any] struct{ A A; B B; C C; D D }
```

Each type has an `Unpack()` method returning the fields as multiple return
values.

### Constructors

```go
T2(a, b)          Tuple2[A, B]
T3(a, b, c)       Tuple3[A, B, C]
T4(a, b, c, d)    Tuple4[A, B, C, D]
```

### Zip

Combines multiple slices into a slice of tuples. Output length equals the
shortest input.

```go
Zip2(a []A, b []B)             []Tuple2[A, B]
Zip3(a []A, b []B, c []C)      []Tuple3[A, B, C]
Zip4(a, b, c, d)               []Tuple4[A, B, C, D]
```

### Unzip

Splits a slice of tuples back into individual slices.

```go
Unzip2(tuples []Tuple2[A, B])          ([]A, []B)
Unzip3(tuples []Tuple3[A, B, C])       ([]A, []B, []C)
Unzip4(tuples []Tuple4[A, B, C, D])    ([]A, []B, []C, []D)
```

### CrossJoin

Produces the Cartesian product of the input slices. Returns an empty slice
immediately if any input is empty.

```go
CrossJoin2(a, b)       []Tuple2[A, B]
CrossJoin3(a, b, c)    []Tuple3[A, B, C]
CrossJoin4(a, b, c, d) []Tuple4[A, B, C, D]
```

### Usage patterns

Tuples are best suited for **local, short-lived grouping within a single
function or pipeline** — situations where defining a one-off named struct would
be more ceremony than it is worth. If the same pairing appears in more than one
place, or is part of a public API, a named struct is almost always clearer.

**1. Pairing a sort key with its original value**

```go
// attach a score without losing the original user
ranked := Map(users, func(u User) Tuple2[int, User] {
    return T2(scoreOf(u), u)
})
```

**2. Zipping parallel slices from separate data sources**

```go
// two separate API responses that logically belong together
rows := Zip2(invoiceIDs, amounts)
ForEach(rows, func(row Tuple2[string, float64]) {
    id, amount := row.Unpack()
    process(id, amount)
})
```

**3. CrossJoin for generating combinations (e.g. test matrix)**

```go
envs    := []string{"staging", "prod"}
regions := []string{"us-east", "eu-west", "ap-south"}

ForEach(CrossJoin2(envs, regions), func(tc Tuple2[string, string]) {
    env, region := tc.Unpack()
    deploy(env, region)
})
```

**4. Unzip to split a parsed CSV into typed columns**

```go
// "alice,30"  "bob,25"  "carol,28"
names, ages := Unzip2(Map(rows, func(row string) Tuple2[string, int] {
    parts := strings.SplitN(row, ",", 2)
    age, _ := strconv.Atoi(parts[1])
    return T2(parts[0], age)
}))
```

**5. Carrying the original alongside a transformed value in a pipeline**

```go
// keep original for later comparison after normalisation
pairs := Map(inputs, func(s string) Tuple2[string, string] {
    return T2(s, normalize(s))
})
changed := Filter(pairs, func(p Tuple2[string, string]) bool {
    return p.A != p.B
})
```

---

## eventbus.go — In-process typed event bus

A lightweight, thread-safe event bus for decoupling components within a single
process. Handlers are registered per event type using Go generics — the type key
is derived via reflection at subscription time, so callers do not need to supply
manual type tokens.

Publishing fans out to all matching handlers concurrently in separate goroutines
and blocks until every handler returns, making `Publish` synchronous from the
caller's perspective.

A `sync.RWMutex` protects the handler map. `Subscribe` takes a write lock while
`Publish` snapshots the relevant handler slice under a read lock and then
releases it before dispatching, so subscriptions from other goroutines are not
blocked by long-running handlers.

Every handler receives the caller's `context.Context`. If the context is already
cancelled when `Publish` is called, no handlers run. Handlers are responsible for
checking `ctx.Err()` themselves if they perform long work — the bus does not
cancel in-flight handlers mid-execution.

### Normal usage

```go
type OrderPlaced struct { OrderID string }
type PaymentFailed struct { Reason string }

bus := NewEventBus()

Subscribe(bus, func(ctx context.Context, e OrderPlaced) {
    fmt.Println("order placed:", e.OrderID)
})

Subscribe(bus, func(ctx context.Context, e PaymentFailed) {
    fmt.Println("payment failed:", e.Reason)
})

Publish(context.Background(), bus, OrderPlaced{OrderID: "ord-123"})
Publish(context.Background(), bus, PaymentFailed{Reason: "insufficient funds"})
```

### Audit/logging handler with `any`

Subscriptions are keyed by the **static generic event type**. A subscription to
`any` is useful for audit/logging paths, but it only receives events that are
published as `Publish[any](...)`.

```go
Subscribe(bus, func(ctx context.Context, e any) {
    switch ev := e.(type) {
    case OrderPlaced:
        log.Printf("audit: order %s placed", ev.OrderID)
    case PaymentFailed:
        log.Printf("audit: payment failed — %s", ev.Reason)
    default:
        log.Printf("audit: event %T", ev)
    }
})

// Must be published as any to match the any subscription key.
Publish[any](context.Background(), bus, OrderPlaced{OrderID: "ord-123"})
```

---

## workerpool.go — Concurrent job processing

Helpers for running work concurrently without managing goroutines directly.
Worker counts less than one are normalized to one.

### `WorkerPool`

A reusable pool that reads jobs from a channel and dispatches them to a fixed
number of goroutines. `Run` blocks until the jobs channel is closed and all
workers have finished.

```go
type HandlerFunc[T any] func(ctx context.Context, job T, logger *slog.Logger)

func NewWorkerPool[T any](numWorkers int, jobs <-chan T, handler HandlerFunc[T], logger *slog.Logger) *WorkerPool[T]
func (p *WorkerPool[T]) Run(ctx context.Context)
```

Use `WorkerPool` when you need control over the jobs channel lifecycle — for
example, when jobs arrive from a long-running stream or message queue:

```go
jobCh := make(chan Order)

go func() {
    for order := range messageQueue.Subscribe() {
        jobCh <- order
    }
    close(jobCh)
}()

handler := func(ctx context.Context, order Order, logger *slog.Logger) {
    if err := processOrder(ctx, order); err != nil {
        logger.Error("order failed", "id", order.ID, "err", err)
    }
}

wp := NewWorkerPool(8, jobCh, handler, slog.Default())
wp.Run(ctx) // blocks until jobCh is closed
```

### `RunConcurrent`

Runs `fn` concurrently over every element of a slice. Blocks until all jobs are
done. Use this when you need side effects but not return values.

```go
func RunConcurrent[T any](ctx context.Context, jobs []T, workerCount int, fn func(context.Context, T))
```

```go
files := []string{"a.csv", "b.csv", "c.csv"}

RunConcurrent(ctx, files, 4, func(ctx context.Context, path string) {
    if err := uploadFile(ctx, path); err != nil {
        log.Printf("upload failed: %s: %v", path, err)
    }
})
```

### `MapConcurrent`

Like `RunConcurrent` but collects results. Output order matches input order
regardless of which worker processes which job.

```go
func MapConcurrent[T any, R any](ctx context.Context, jobs []T, workerCount int, fn func(context.Context, T) R) []R
```

```go
userIDs := []int{1, 2, 3, 4, 5}

profiles := MapConcurrent(ctx, userIDs, 4, func(ctx context.Context, id int) Profile {
    p, err := fetchProfile(ctx, id)
    if err != nil {
        return Profile{} // zero value on failure
    }
    return p
})
// profiles[0] corresponds to userIDs[0], profiles[1] to userIDs[1], etc.
```

---

## Related iterator tooling

This package intentionally keeps `slices.go` eager and dependency-free. For lazy
iterator evaluation and composable pipelines, consider using Go's standard
`iter` package directly or
[`github.com/BooleanCat/go-functional`](https://github.com/BooleanCat/go-functional).
