# PAS: Promise, Async, Sync

PAS is a Go package that provides a simple and efficient way to handle asynchronous computations using Promises. By simply wrapping your functions in `pas.Async[T](yourFunction, args...)`, you can instantly make them parallel, without any extra code.

Promises are resolved automatically when they are used as arguments to other PAS-wrapped functions. Promises that are nested arbitrarily deep inside slices, maps, or pointers are also resolved automatically.

```go
// 🔄 Sequential
array := make([]int, 100)
for i := range array {
    if i <= 50 {
        array[i] = Compute(i, 0) // Some intensive computation
    } else {
        array[i] = Compute(i, array[i-50]) // Complex inter-dependency
    }
}
sum := Sum(array)
fmt.Println(sum)
```

⬇️ ⬇️ ⬇️

```go
// ⚡️ Parallel
array := pas.MakeSlice[int](10) // or: make([]*pas.Promise[int], 10)
for i := range array {
    if i <= 5 {
        // No need to change implementation of Compute
        array[i] = pas.Async[int](Compute, i, 0)
    } else {
        // Dependencies are automatically handled
        array[i] = pas.Async[int](Compute, i, array[i-5])
    }
}
// Promised values inside slices/maps/pointers are automatically resolved
sum := pas.Sync[int](Sum, array)
fmt.Println(sum)
```

## Table of Contents

- [PAS: Promise, Async, Sync](#pas-promise-async-sync)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Creating Promises](#creating-promises)
    - [Asynchronous Operations](#asynchronous-operations)
    - [Synchronous Operations](#synchronous-operations)
    - [Example](#example)
  - [API Reference](#api-reference)
    - [`Promise`](#promise)
    - [`Promise.Get`](#promiseget)
    - [`New`](#new)
    - [`Async`](#async)
    - [`Sync`](#sync)
    - [`MakeSlice`](#makeslice)
    - [`MakeMap`](#makemap)
  - [Limitations](#limitations)
  - [Implementation Details](#implementation-details)
  - [License](#license)

## Features

- **Promises:** Represent future values, allowing you to handle results of asynchronous operations.
- **Async:** Execute functions in parallel with automatic handling of dependent (even nested) Promises.
- **Sync:** Execute functions synchronously with automatic handling of dependent (even nested) Promises.

## Installation

To install the latest version of PAS package, use `go get`:

```bash
go get github.com/AnJunHao/pas@latest
```

## Usage

### Creating Promises

Create a new Promise using the `New` function. The Promise is immediately ready. You can initialize it with an optional value.

```go
import "github.com/AnJunHao/pas"

// Create a promise with an initial value
p := pas.New(42)

// Create a promise with its initial value as the default value of type T
p := pas.New[int]()
```

### Asynchronous Operations

Use the `Async` function to execute a function in parallel. It takes a function and its arguments, which can include other Promises or nested Promises. Dependent Promises or nested Promises are resolved automatically.

```go
import "github.com/AnJunHao/pas"

func Compute(x, y int) int {
    // Some intensive computation
    return x + y
}

p := pas.Async[int](Compute, 5, 10)
result := p.Get() // Blocks until the computation is done
```

### Synchronous Operations

Use the `Sync` function to execute a function synchronously. It automatically waits for any Promises or nested Promises passed as arguments to be resolved.

```go
import "github.com/AnJunHao/pas"

func Multiply(x, y int) int {
    return x * y
}

p1 := pas.Async[int](Compute, 5, 10)
result := pas.Sync[int](Multiply, p1, 3)
```

### Example

Here's a complete example demonstrating parallel sum computation:

```go
// main.go
package main

import (
    "fmt"
    "time"

    "github.com/AnJunHao/pas"
)

func SumWithinRange(start int, end int) int {
    sum := 0
    for i := start; i <= end; i++ {
        sum += i
    }
    return sum
}

func Add(a, b int) int {
    return a + b
}

func main() {
    n := 1000000000
    numWorkers := 20

    startTime := time.Now()

    // Parallel Sum
    parSum := pas.New(0)
    for i := 0; i < numWorkers; i++ {
        start := i*n/numWorkers + 1
        end := (i + 1) * n / numWorkers
        s := pas.Async[int](SumWithinRange, start, end)
        parSum = pas.Async[int](Add, parSum, s)
    }

    // Get the result
    fmt.Println(parSum.Get())
    duration := time.Since(startTime)
    fmt.Printf("Parallel Sum took: %v\n", duration)

    // Sequential Sum
    startTime = time.Now()
    seqSum := SumWithinRange(0, n)
    duration = time.Since(startTime)
    fmt.Printf("Sequential Sum took: %v\n", duration)

    if seqSum == parSum.Get() {
        fmt.Println("Success: Sequential and Parallel results match.")
    } else {
        fmt.Println("Error: Results do not match.")
    }
}
```

## API Reference

### `Promise`

Represents a parallel variable holding a value of type `T`.

```go
type Promise[T any] struct {
    value T
    ready chan struct{}
    once  sync.Once
}
```

### `Promise.Get`

Returns the computed value, blocking until it is ready.

```go
func (p *Promise[T]) Get() T
```

### `New`

Creates a pointer to a new Promise with an optional initial value. The Promise is immediately ready.

```go
func New[T any](value ...T) *Promise[T]
```

### `Async`

Starts a parallel computation by invoking function `f` with the provided arguments. If any argument is a Promise, it waits for it to be ready before executing `f`.

```go
func Async[T any](f interface{}, args ...interface{}) *Promise[T]
```

**Parameters:**

- `f`: The function to execute asynchronously. It must have exactly one return value of type `T`.
- `args`: Arguments to pass to the function. Can include Promises.

**Returns:**

- `*Promise[T]`: A Promise representing the future result of the computation.

### `Sync`

Executes function `f` synchronously with the provided arguments. If any argument is a Promise, it waits for it to be ready before executing `f`.

```go
func Sync[T any](f interface{}, args ...interface{}) T
```

**Parameters:**

- `f`: The function to execute synchronously. It must have exactly one return value of type `T`.
- `args`: Arguments to pass to the function. Can include Promises.

**Returns:**

- `T`: The result of the function execution.

### `MakeSlice`

Creates a slice of `*Promise[T]` with the specified length and capacity. The Promises are immediately ready.

```go
func MakeSlice[T any](length int, capacity ...int) []*Promise[T]
```

**Parameters:**

- `length`: The number of elements in the slice.
- `capacity` (optional): The capacity of the slice.

**Returns:**

- `[]*Promise[T]`: A slice of Promises.

**Usage:**

```go
promises := pas.MakeSlice[int](5)
```

### `MakeMap`

Creates a map with keys of type `K` and values of type `*Promise[V]`.

```go
func MakeMap[K comparable, V any](size ...int) map[K]*Promise[V]
```

**Parameters:**

- `size` (optional): Initial size of the map.

**Returns:**

- `map[K]*Promise[V]`: A map with Promises as values.

**Usage:**

```go
promiseMap := pas.MakeMap[string, int](10)
```

## Limitations

- `Async` and `Sync` only work with functions that **return exactly one** value.
- `Async` and `Sync` do not work with methods (functions with a receiver).
- `Async` and `Sync` do not work with variadic functions (functions with a variable number of arguments).
- No pooling mechanism is implemented (yet). Each call to `Async` creates a new goroutine.

## Implementation Details

- `Async` and `Sync` calls an internal function `executeFunction` that handles the details of resolving arguments and calling the function.
- `executeFunction` uses reflection to inspect the type and value of each argument, and calls `resolveValue` to resolve Promises and nested Promises recursively.
- `resolveValue` recursively resolves Promises within the input based on the expected type. It handles Promises, pointers, slices, arrays, maps, and nested combinations thereof. Non-Promise arguments are returned as-is.
  - Example inputs and expected outputs:
    - `*Promise[int]` -> `int`
    - `[]*Promise[int]` -> `[]int`
    - `[]map[string]*Promise[[]map[string]int]` -> `[]map[string][]map[string]int`

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
