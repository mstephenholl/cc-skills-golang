# Pointer Types Deep Dive

## Table of Contents

- [Regular Pointers (`*T`)](#regular-pointers-t)
  - [Stack vs Heap (Escape Analysis)](#stack-vs-heap-escape-analysis)
  - [`new(T)` vs `&T{}`](#newt-vs-t)
- [`unsafe.Pointer`](#unsafepointer)
  - [The 6 Valid Patterns (from the `unsafe` package docs)](#the-6-valid-patterns-from-the-unsafe-package-docs)
  - [Critical Rule: NEVER Store `uintptr` Across Statements](#critical-rule-never-store-uintptr-across-statements)
  - [Modern Alternatives (prefer these)](#modern-alternatives-prefer-these)
- [`weak.Pointer[T]` (Go 1.24+)](#weakpointert-go-124)
  - [Use Cases](#use-cases)
  - [`runtime.AddCleanup` vs `runtime.SetFinalizer`](#runtimeaddcleanup-vs-runtimesetfinalizer)

## Regular Pointers (`*T`)

### Stack vs Heap (Escape Analysis)

Go's compiler decides whether to allocate on the stack or heap. A variable "escapes" to the heap when its lifetime extends beyond the function:

```go
func noEscape() int {
    x := 42
    return x // x stays on stack — copied on return
}

func escapes() *int {
    x := 42
    return &x // x escapes to heap — pointer outlives function
}
```

Use `go build -gcflags="-m"` to see escape analysis decisions. Heap allocations add GC pressure — avoid unnecessary escapes in hot paths.

### `new(T)` vs `&T{}`

Both allocate and return a pointer. `&T{}` is preferred because it allows field initialization:

```go
p := new(Point)       // *Point with zero values
p := &Point{X: 1}     // *Point with initialized fields — preferred
```

## `unsafe.Pointer`

`unsafe.Pointer` bypasses Go's type system for FFI and low-level memory manipulation. Only the 6 patterns documented in the `unsafe` package are valid; any other pattern is undefined behavior.

### The 6 Valid Patterns (from the `unsafe` package docs)

**Pattern 1: Convert `*T` to `*U` via `unsafe.Pointer`**

```go
// Reinterpret a float64 as its raw bits
f := 1.5
bits := *(*uint64)(unsafe.Pointer(&f))
```

**Pattern 2: Convert `unsafe.Pointer` to `uintptr` (but not back)** — the integer is only for printing or hashing an address; it keeps nothing alive.

**Pattern 3: Convert `unsafe.Pointer` to `uintptr` and back, with arithmetic, in one expression**

```go
// Pointer arithmetic — MUST be a single expression; prefer unsafe.Add (below)
p := unsafe.Pointer(uintptr(unsafe.Pointer(&s.field)) + offset)
```

**Pattern 4: `syscall.Syscall` arguments** — the conversion must appear in the call expression itself.

```go
syscall.Syscall(SYS_READ, fd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
```

**Pattern 5: `reflect.Value.Pointer()` or `UnsafeAddr()` to `unsafe.Pointer`, immediately**

```go
p := unsafe.Pointer(reflect.ValueOf(&x).Pointer())
```

**Pattern 6: `reflect.SliceHeader`/`StringHeader` `Data` field to or from `unsafe.Pointer`** — both headers are deprecated; use `unsafe.Slice`, `unsafe.String`, `unsafe.SliceData` and `unsafe.StringData` instead.

### Critical Rule: NEVER Store `uintptr` Across Statements

```go
// ✗ DANGEROUS — a uintptr is not a reference: between these two lines the GC
// may free x, or stack growth may move it
u := uintptr(unsafe.Pointer(&x))
p := unsafe.Pointer(u) // may now be dangling

// ✓ Safe — single expression
p := unsafe.Pointer(uintptr(unsafe.Pointer(&x)) + offset)
```

### Modern Alternatives (prefer these)

| Function | Since | Purpose |
| --- | --- | --- |
| `unsafe.Add(ptr, len)` | Go 1.17 | Pointer arithmetic without `uintptr` conversion |
| `unsafe.Slice(ptr, len)` | Go 1.17 | Create slice from pointer + length |
| `unsafe.String(ptr, len)` | Go 1.20 | Create string from pointer + length |
| `unsafe.SliceData(s)` | Go 1.17 | Get pointer to slice's backing array |
| `unsafe.StringData(s)` | Go 1.20 | Get pointer to string's backing array |

These are safer than manual `uintptr` arithmetic because they keep values as pointers (visible to GC) throughout.

## `weak.Pointer[T]` (Go 1.24+)

A weak pointer holds a reference to an object without preventing garbage collection. When the GC reclaims the object, `Value()` returns `nil`.

```go
strong := new(MyType)
w := weak.Make(strong)

if p := w.Value(); p != nil {
    // object still alive
} else {
    // object was garbage collected
}
```

### Use Cases

- **Canonicalization / interning** — deduplicate equivalent values without keeping them alive; pair with `runtime.AddCleanup` to delete the dead map entry
- **Side tables keyed by object lifetime** — metadata that should disappear with the object it describes

Not for caches that must retain entries: a weak entry vanishes as soon as no strong reference remains, so hit rates collapse under GC.

### `runtime.AddCleanup` vs `runtime.SetFinalizer`

Prefer `runtime.AddCleanup` (Go 1.24+) over `runtime.SetFinalizer`:

- Multiple cleanups can be registered per object
- Cleanup function receives a value, not a pointer to the collected object
- No risk of resurrecting the object
- Works correctly with weak pointers
