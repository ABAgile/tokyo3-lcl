# lcl - the "water of life" for mental synchronization

[![Release](https://img.shields.io/github/v/release/abagile/tokyo3-lcl?sort=semver&logo=Go&color=%23007D9C)](https://github.com/abagile/tokyo3-lcl/releases)
[![Test](https://github.com/abagile/tokyo3-lcl/actions/workflows/test.yml/badge.svg)](https://github.com/abagile/tokyo3-lcl/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/abagile/tokyo3-lcl.svg)](https://pkg.go.dev/github.com/abagile/tokyo3-lcl)
[![Go Report Card](https://goreportcard.com/badge/github.com/abagile/tokyo3-lcl)](https://goreportcard.com/report/github.com/abagile/tokyo3-lcl)
[![codecov](https://codecov.io/gh/abagile/tokyo3-lcl/branch/main/graph/badge.svg)](https://codecov.io/gh/abagile/tokyo3-lcl)

A generic Go utility library providing useful helpers for basic yet handy
operations, as the core building block for most applications & utility
libraries

**Requires Go 1.26+**

```
go get github.com/abagile/tokyo-lcl
```

---

## Design philosophy

### Iterator-first collection manipulation

All slice operations in `slices.go` are built on Go 1.23's `iter.Seq` / `iter.Seq2`
push-iterator model rather than operating on concrete slices internally. Functions
accept a `[]T` for ergonomics at the call site, but immediately convert to iterators
via `slices.Values` / `slices.All` and delegate the actual traversal to
[`github.com/BooleanCat/go-functional`](https://github.com/BooleanCat/go-functional).

`iter.Seq[V]` is defined as `func(yield func(V) bool)` — the iterator owns the
loop and pushes each value into the consumer's `yield` callback (an internal
iterator). This is distinct from a pull model, where the consumer calls `next()` to
request one value at a time; Go's `iter.Pull` can convert a push iterator into that
style when needed, but ranging directly over an `iter.Seq` stays in push mode and is
the idiomatic choice here.

This means:
- **Composition is lazy by default** — `Filter`, `Map`, etc. chain without allocating intermediate slices inside the pipeline.
- **`Error` variants short-circuit cleanly** — iteration stops the moment a predicate or mapper returns an error, with no extra bookkeeping in the caller.
- **Custom slice types are preserved** — all functions are constrained on `S ~[]T`, so a named type such as `type UserList []User` comes out the other side unchanged.

The iterator model is most valuable for operations that can terminate early
— `Filter`, `Exclude`, `MapError`, `FoldError`, `ForEachWhile`, and similar.
For operations that must consume the entire input before producing a result,
the benefit is marginal. `Mean` and `Mode` fall into this category: they
require a full pass regardless, so they accept a concrete `[]T` instead, which
lets them chain directly with the output of any `slices.go` helper and makes
the full-consumption contract explicit at the call site.

---

## slices.go — Functional slice operations

All functions preserve the element type of custom slice types (`S ~[]T`).

### Function naming conventions

Each operation comes in up to three variants that mirror Go's iterator types directly:

| Suffix | Iterator type | Meaning |
|---|---|---|
| *(none)* | `iter.Seq[V]` | Plain value sequence |
| `2` | `iter.Seq2[K, V]` | `(index, value)` pair — index is the slice position |
| `Error` | — | Same as the plain variant, but the predicate/mapper may return an `error`; iteration stops immediately on the first non-nil error |

So `Filter`, `Filter2`, and `FilterError` are the same operation expressed at three
levels of richness. Reach for the plain variant by default; add `2` when you need
the position, add `Error` when the callback can fail.

These helpers cover the common case of working with a concrete `[]T`. They are
intentionally thin — each one converts the slice to an iterator, delegates to
[`go-functional/it`](https://github.com/BooleanCat/go-functional), and collects the
result back into a slice. For pipelines that compose multiple steps, operate on
non-slice sources, or need finer control over iteration, use the `it` package
directly rather than chaining these helpers.

### Filtering
| Function | Description |
|---|---|
| `Filter(xs, pred)` | Keep elements where `pred` is true |
| `Filter2(xs, pred)` | Like `Filter` but predicate receives `(index, value)` |
| `FilterError(xs, pred)` | Like `Filter` but predicate may return an error; stops on first error |
| `Exclude(xs, pred)` | Keep elements where `pred` is false |
| `Exclude2(xs, pred)` | Like `Exclude` with index-aware predicate |
| `ExcludeError(xs, pred)` | Like `Exclude` but predicate may return an error |

### Transformation
| Function | Description |
|---|---|
| `Map(xs, mapper)` | Transform each element |
| `Map2(xs, mapper)` | Transform with `(index, value)` mapper |
| `MapError(xs, mapper)` | Transform; stops on first error |

### Folding
| Function | Description |
|---|---|
| `Fold(xs, accum, initial)` | Left fold |
| `Fold2(xs, accum, initial)` | Left fold with index |
| `FoldError(xs, accum, initial)` | Left fold; stops on first error |
| `FoldRight(xs, accum, initial)` | Right fold |
| `FoldRight2(xs, accum, initial)` | Right fold with index |
| `FoldRightError(xs, accum, initial)` | Right fold; stops on first error |

### Iteration
| Function | Description |
|---|---|
| `ForEach(xs, f)` | Call `f` for every element |
| `ForEach2(xs, f)` | Call `f(index, value)` for every element |
| `ForEachWhile(xs, pred)` | Iterate while `pred` returns true |

### Grouping
| Function | Description |
|---|---|
| `GroupBy(xs, mapper)` | `map[R]S` — groups share the same key; order within each group is preserved |
| `PartitionBy(xs, mapper)` | `[]S` — like `GroupBy` but returns an ordered slice of groups in first-seen key order, rather than a map |
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

---

## result.go — Panic-on-failure monadic helper

Designed for CLI tools and startup paths where certain failures have no sensible
recovery action — a missing required environment variable, a database connection
that must succeed, a NATS dial that is a hard dependency. Rather than threading
errors through every call frame, wrap the fallible operation and call a `Must*`
method: the process panics with a clear, formatted message on failure, or you get
the value and move on.

Two types cover distinct cases:

| Type | Carries | Use when |
|---|---|---|
| `Result[T]` | value only | you already have the value; no error path |
| `ResultE[T]` | value + error | wrapping a fallible operation |

### Constructors

| Constructor | Returns | Description |
|---|---|---|
| `ResultVal(v)` | `*Result[T]` | Wraps a plain value |
| `ResultOf(v, err)` | `*ResultE[T]` | Wraps any `(T, error)` pair |

### Methods

**`Result[T]`**

```go
func (r *Result[T]) Val() T
func (r *Result[T]) MustPass(msg string, v ...any)    // when T is error: panics if non-nil
func (r *Result[T]) MustPresent(msg string, v ...any) // panics if value is zero
```

**`ResultE[T]`**

```go
func (r *ResultE[T]) Val() T
func (r *ResultE[T]) Err() error
func (r *ResultE[T]) Unwrap() (T, error)
func (r *ResultE[T]) Bind(f func(T) *ResultE[T]) *ResultE[T]
func (r *ResultE[T]) MustPass(msg string, v ...any)    // panics if err != nil
func (r *ResultE[T]) MustGet(msg string, v ...any) T   // MustPass + Val
func (r *ResultE[T]) MustPresent(msg string, v ...any) // MustPass + zero-value check
```

- **`Bind`** — chains operations on `ResultE`; short-circuits on error, making it easy to compose multiple fallible steps before the final `Must*` call.
- **`MustPass`** — panics with `"<msg>: <err>"` if the result holds an error. On `Result[T]`, only applicable when `T` is `error`.
- **`MustGet`** — panics on error, otherwise returns the value.
- **`MustPresent`** — panics on error or zero value; use when a non-nil, non-zero result is required.

### Examples

**`ResultOf` + `MustGet`** — wrap any `(T, error)` return and extract the value:

```go
dsn := ResultOf(os.LookupEnv("DATABASE_URL")).MustGet("DATABASE_URL is required")
db  := ResultOf(sql.Open("pgx", dsn)).MustGet("failed to open database: %s", dsn)
```

**`ResultVal` + `MustPresent`** — assert a value is non-zero when there is no error path:

```go
// os.Getenv returns "" on missing — no error, but empty is still wrong
ResultVal(os.Getenv("REDIS_HOST")).MustPresent("REDIS_HOST must not be empty")
ResultVal(cfg.NATSUrl).MustPresent("nats_url is required in config")
```

**`ResultVal` + `MustPass`** — when the value itself is an `error`:

```go
ResultVal(db.PingContext(ctx)).MustPass("database ping failed")
ResultVal(os.MkdirAll(cfg.DataDir, 0755)).MustPass("failed to create data dir %q", cfg.DataDir)
```

**`ResultOf` + `MustPresent`** — assert both success and a non-zero value in one call:

```go
user := ResultOf(db.FindUser(ctx, id)).MustPresent("user %d not found", id)
```

**`Bind`** — chain multiple fallible steps before the final assertion:

```go
token := ResultOf(os.LookupEnv("API_TOKEN")).
    Bind(func(t string) *ResultE[string] {
        if len(t) < 32 {
            return ResultOf("", fmt.Errorf("token too short"))
        }
        return ResultOf(strings.TrimSpace(t), nil)
    }).
    MustGet("invalid API_TOKEN")
```

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
slice. Accepts the result of any `slices.go` helper directly.

### `Mode`
```go
func Mode[T comparable](xs []T) ([]T, bool)
```
Returns all values that appear with the highest frequency. Returns `([],
false)` for an empty slice. Order of results is non-deterministic when multiple
values tie. Accepts the result of any `slices.go` helper directly.

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
func ToPtr[T comparable](v T) *T                      // nil if v is zero
func FromPtr[T any](ptr *T, fallback ...T) T          // dereference or fallback/zero
```

### Slice conversion

```go
func ToAnySlice[T any](in []T) []any
func FromAnySlice[T any](in []any) ([]T, bool)   // false if any element cannot be asserted to T
```

### Nested data access

`GetIn` and `SetIn` navigate nested `map[string]any`, `map[any]any`, and `[]any` structures
using a dot-separated path string. Slice elements are accessed by numeric index.

```go
func GetIn(data any, path string) (any, error)
func SetIn(data any, path string, value any) error
```

```go
data := map[string]any{
    "user": map[string]any{
        "scores": []any{10, 20, 30},
    },
}

v, _ := GetIn(data, "user.scores.1")  // 20
SetIn(data, "user.scores.1", 99)      // scores[1] = 99
```

`SetIn` automatically creates intermediate `map[string]any` nodes for missing keys.
It does **not** grow slices — the index must already be in bounds.

### Identity

```go
func Id[T any](v T) T
```

Returns its argument unchanged. Primarily useful as a first-class mapper when the
element itself should be the key — avoiding the need to write `func(x T) T { return x }`
at every call site:

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

Each type has an `Unpack()` method returning the fields as multiple return values.

### Constructors

```go
T2(a, b)          Tuple2[A, B]
T3(a, b, c)       Tuple3[A, B, C]
T4(a, b, c, d)    Tuple4[A, B, C, D]
```

### Zip

Combines multiple slices into a slice of tuples. Output length equals the
longest input; shorter inputs are padded with zero values.

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
process. Handlers are registered per event type using Go generics — the type
key is derived via reflection at subscription time, so the map stays strongly
typed without requiring callers to supply type tokens manually. Publishing fans
out to all matching handlers concurrently in separate goroutines and blocks
until every handler returns, making `Publish` synchronous from the caller's
perspective.

A `sync.RWMutex` protects the handler map; `Subscribe` takes a write lock while
`Publish` snapshots the relevant handler slice under a read lock and then
releases it before dispatching, so subscriptions from other goroutines are
never blocked by long-running handlers.

Every handler receives the caller's `context.Context`. If the context is
already cancelled when `Publish` is called, no handlers run. Handlers are
responsible for checking `ctx.Err()` themselves if they perform long work — the
bus does not cancel in-flight handlers mid-execution.

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

### Catch-all handler

Subscribe to `any` to receive every event regardless of type — useful for
logging, tracing, or forwarding to an external sink:

```go
Subscribe(bus, func(ctx context.Context, e any) {
    switch ev := e.(type) {
    case OrderPlaced:
        log.Printf("audit: order %s placed", ev.OrderID)
    case PaymentFailed:
        log.Printf("audit: payment failed — %s", ev.Reason)
    }
})

// Must be published as any to match the any subscription key
Publish[any](context.Background(), bus, OrderPlaced{OrderID: "ord-123"})
```

---

## workerpool.go — Concurrent job processing

Helpers for running work concurrently without managing goroutines directly.

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
func RunConcurrent[T any](jobs []T, workerCount int, fn func(context.Context, T))
```

```go
files := []string{"a.csv", "b.csv", "c.csv"}

RunConcurrent(files, 4, func(ctx context.Context, path string) {
    if err := uploadFile(ctx, path); err != nil {
        log.Printf("upload failed: %s: %v", path, err)
    }
})
```

### `MapConcurrent`

Like `RunConcurrent` but collects results. Output order matches input order
regardless of which worker processes which job.

```go
func MapConcurrent[T any, R any](jobs []T, workerCount int, fn func(context.Context, T) R) []R
```

```go
userIDs := []int{1, 2, 3, 4, 5}

profiles := MapConcurrent(userIDs, 4, func(ctx context.Context, id int) Profile {
    p, err := fetchProfile(ctx, id)
    if err != nil {
        return Profile{} // zero value on failure
    }
    return p
})
// profiles[0] corresponds to userIDs[0], profiles[1] to userIDs[1], etc.
```

---

## Acknowledgements

The slice and iterator utilities in this library are built on top of
[**BooleanCat/go-functional**](https://github.com/BooleanCat/go-functional), which
provides the composable `iter.Seq`-based primitives (`it.Filter`, `it.Map`,
`it.Fold`, `it.ForEach`, and their `2` / `Error` variants) that power the
collection functions here. go-functional deserves full credit for the iterator
plumbing that makes lazy, allocation-free pipelines possible in idiomatic Go.
