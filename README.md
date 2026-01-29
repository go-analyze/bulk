# go-analyze/bulk

[![license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/go-analyze/bulk/blob/master/LICENSE)
[![Build Status](https://github.com/go-analyze/bulk/actions/workflows/tests-main.yml/badge.svg)](https://github.com/go-analyze/bulk/actions/workflows/tests-main.yml)

**Performance-first collection operations for Go**

`bulk` provides optimized utilities for working with slices and maps in Go, designed to **minimize memory allocations** and **maximize performance**. Unlike typical collection libraries that always copy, copies are avoided when possible (instead using views or truncated slices). And for uses where the input view can be discarded, `InPlace` functionality is offered to prevent copying entirely.

---

## Core Patterns

### Function Variants
- **Base functions** (e.g., `SliceFilter`): Optimized operations that avoid allocations when possible (views returned)
- **`InPlace` variants** (e.g., `SliceFilterInPlace`): Zero-allocation operations that modify input slices
- **`Into` variants** (e.g., `SliceFilterInto`): Append results to existing slices
- **`By` variants** (e.g., `SliceToGroupsBy`): Use custom key functions for operations

### Performance Strategy
- Default functions return views or truncated slices when safe
- Multi-slice operations enable cross-slice optimizations
- `InPlace` variants provide zero-allocation guarantees for performance-critical code

### Important: Memory Safety
**Default functions may return views** - appending to results may affect the original slice:
```go
data := []int{1, 2, 4, 5, 10, 15, 20}
lowVals := bulk.SliceFilter(func(n int) bool { return n <= 4 }, data)
// lowVals may be a view of data

// Safe: read-only usage
fmt.Println(lowVals)

// Unsafe: modifying the result
evens = append(lowVals, 99) // ⚠️ May corrupt original data slice
```

**When to use each variant:**
- **Default functions**: When you won't modify the result (read-only usage)
- **`InPlace` variants**: When input slice can be discarded after operation
- **Copy-safe alternatives**: Consider [lo](https://github.com/samber/lo), [Pie](https://github.com/elliotchance/pie), or a manual copy on result from `bulk`

---

## Installation

```bash
go get github.com/go-analyze/bulk
```

Import in your code:

```go
import "github.com/go-analyze/bulk"
```

---

## Essential Operations

### Filtering & Processing

**`SliceFilter[T any](predicate func(v T) bool, slices ...[]T) []T`**  
Filters elements that pass the predicate function from one or multiple slices. **Zero allocations** when all true elements are consecutive.

```go
numbers := []int{1, 2, 3, 4, 5, 6}
evens := bulk.SliceFilter(func(n int) bool { return n%2 == 0 }, numbers)
// Result: [2, 4, 6]
// Original slice unchanged: [1, 2, 3, 4, 5, 6]
```

**`SliceFilterInPlace[T any](predicate func(v T) bool, slices ...[]T) []T`**  
**Zero-allocation** filtering by reusing input slice memory (zero allocation only guaranteed with a single slice, but will always have less allocations than SliceFilter).

```go
numbers := []int{1, 2, 3, 4, 5, 6} // This slice will be modified!
evens := bulk.SliceFilterInPlace(func(n int) bool { return n%2 == 0 }, numbers)
// Result: [2, 4, 6]
// Warning: numbers slice is now corrupted and must be discarded
```

**`SliceFilterTransform[I, R any](predicate func(I) bool, transform func(I) R, inputs ...[]I) []R`**  
Filter and convert in one pass - the most efficient pattern for common data processing.

```go
numbers := []int{1, 2, 3, 4, 5, 6}
evenStrings := bulk.SliceFilterTransform(
    func(n int) bool { return n%2 == 0 },     // filter: keep evens
    func(n int) string { return fmt.Sprintf("num_%d", n) }, // transform: to strings
    numbers)
// Result: ["num_2", "num_4", "num_6"]
```

### Partitioning

**`SliceSplit[T any](predicate func(v T) bool, slices ...[]T) ([]T, []T)`**  
Partitions elements from one or multiple slices into two slices based on predicate.

```go
numbers := []int{1, 2, 3, 4, 5, 6}
evens, odds := bulk.SliceSplit(func(n int) bool { return n%2 == 0 }, numbers)
// evens: [2, 4, 6], odds: [1, 3, 5]
```

**`SliceSplitInPlace` and `SliceSplitInPlaceUnstable` variants** provide memory-efficient and fastest partitioning respectively.

### Set Operations

**`SliceToSet[T comparable](slices ...[]T) map[T]struct{}`**  
Converts slices to a set for fast lookup and deduplication. Accepts multiple slices for union operations.

```go
slice1 := []string{"a", "b", "c", "b"}
slice2 := []string{"c", "d", "e"}
set := bulk.SliceToSet(slice1, slice2)
// Result: map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}, "e": {}}

// Deduplication pattern:
duplicates := []string{"apple", "banana", "apple", "cherry", "banana"}
unique := bulk.MapKeySlice(bulk.SliceToSet(duplicates))
// Result: ["apple", "banana", "cherry"] (order may vary)
```

**`SliceToSetBy[I any, R comparable](keyfunc func(I) R, slices ...[]I) map[R]struct{}`**  
Creates a set using a key function to transform elements into comparable keys. Typically used to provide a field from within the structs within the slice.

**`SliceIntersect[T comparable](a, b []T) []T`**  
Returns elements that exist in both slices, preserving order from slice a. Duplicates are automatically removed from the result.

```go
userA := []string{"music", "sports", "reading", "cooking"}
userB := []string{"sports", "movies", "cooking", "travel"}
common := bulk.SliceIntersect(userA, userB)
// Result: ["sports", "cooking"]
```

**`SliceDifference[T comparable](a, b []T) []T`**  
Returns elements that exist in slice a but not in slice b, preserving order from slice a. Duplicates are automatically removed from the result.

```go
userA := []string{"music", "sports", "reading", "cooking"}
userB := []string{"sports", "movies", "cooking", "travel"}
unique := bulk.SliceDifference(userA, userB)
// Result: ["music", "reading"] - things userA likes that userB doesn't
```

### Data Organization

**`SliceToCounts[T comparable](slices ...[]T) map[T]int`**  
Counts occurrences of each element across multiple slices.

```go
slice1 := []string{"a", "b", "c", "b"}
slice2 := []string{"c", "d", "a"}
counts := bulk.SliceToCounts(slice1, slice2)
// Result: map[string]int{"a": 2, "b": 2, "c": 2, "d": 1}
```

**`SliceToIndexBy[T any, K comparable](keyfunc func(T) K, slices ...[]T) map[K]T`**  
Creates an index map where each key maps to the **last** value encountered.

```go
type Person struct{ ID int; Name string }
people := []Person{{1, "Alice"}, {2, "Bob"}, {1, "Alice_Updated"}}
index := bulk.SliceToIndexBy(func(p Person) int { return p.ID }, people)
// Result: map[int]Person{1: {1, "Alice_Updated"}, 2: {2, "Bob"}} (last wins)
```

**`SliceToGroupsBy[T any, K comparable](keyfunc func(T) K, slices ...[]T) map[K][]T`**  
Groups elements by key derived by each entry, preserving all values for each key.

```go
type Person struct{ Dept, Name string }
people := []Person{
    {"eng", "Alice"}, {"sales", "Bob"}, 
    {"eng", "Charlie"}, {"sales", "Dave"},
}
groups := bulk.SliceToGroupsBy(func(p Person) string { return p.Dept }, people)
// Result: map[string][]Person{
//   "eng": [{"eng", "Alice"}, {"eng", "Charlie"}],
//   "sales": [{"sales", "Bob"}, {"sales", "Dave"}]
// }
```

---

## Function Discovery

**Naming Pattern**: `[Collection][Operation][Variant]`

### Collections
- **`Slice`**: Operations on slices (e.g., `SliceFilter`, `SliceToSet`)
- **`Map`**: Operations on maps (e.g., `MapInvert`)

### Common Variants
- **`InPlace`**: Zero-allocation, modifies input (e.g., `SliceFilterInPlace`)
- **`Into`**: Append to existing collection (e.g., `SliceFilterInto`, `SliceIntoSet`)
- **`By`**: Use custom key function (e.g., `SliceToSetBy`, `SliceToGroupsBy`)

### Error Handling
- **`Err`**: Functions that can return errors (e.g., `SliceFilterTransformErr`)

Many operations offer multiple variants - check function signatures for the complete API.

---

## When to use `bulk`

* **Large data analyses** where minimizing memory pressure is critical
* **Performance-sensitive loops** processing millions of elements
* Scenarios where **in-place mutations** are safe and desired (see `Memory Safety` above)

If you require copy-on-write semantics as your primary workflow, consider the other options described in `Memory Safety`.

## Project Versioning

This project follows [ZeroVer](https://0ver.org) versioning, and will never release a v1. Instead, middle digit increases (e.g., `v0.1.X` to `v0.2.0`) indicate breaking changes such as Go version updates or API modifications. Patch version increases (e.g., `v0.2.0` to `v0.2.1`) indicate bug fixes and feature expansions, which are generally safe to apply.

