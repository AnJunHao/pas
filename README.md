# PAS: Promise, Async, Sync

PAS is a Go package that provides a simple and efficient way to handle asynchronous computations using Promises. By simply wrapping your functions in `pas.Async[T](yourFunction, args...)`, you can instantly make them parallel, without any extra code.

```go
// 🔄 Sequential
accumulator := -5
intermediate := Compute(5, 10)
accumulator = accumulator + intermediate
result := Calculate(accumulator, 15)
fmt.Println(result)
```

⬇️ ⬇️ ⬇️

```go
// ⚡️ Parallel
accumulatorP := pas.New(-5)
intermediateP := pas.Async[int](Compute, 5, 10)
accumulatorP = pas.Async[int](Add, accumulatorP, intermediateP)
resultP := pas.Async[int](Calculate, accumulatorP, 15) // Dependencies are automatically handled
fmt.Println(resultP.Get())
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
    - [Examples](#examples)
  - [API Reference](#api-reference)
    - [`Promise`](#promise)
    - [`New`](#new)
    - [`Async`](#async)
    - [`Sync`](#sync)
    - [`MakeSlice`](#makeslice)
    - [`MakeMap`](#makemap)
    - [`MakeChan`](#makechan)
  - [License](#license)

## Features

- **Promises:** Represent future values, allowing you to handle results of asynchronous operations.
- **Async:** Execute functions in parallel with automatic handling of dependent Promises.
- **Sync:** Execute functions synchronously with automatic handling of dependent Promises.

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

Use the `Async` function to execute a function in parallel. It takes a function and its arguments, which can include other Promises.

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

Use the `Sync` function to execute a function synchronously. It automatically waits for any Promises passed as arguments to be resolved.

```go
import "github.com/AnJunHao/pas"

func Multiply(x, y int) int {
    return x * y
}

p1 := pas.Async[int](Compute, 5, 10)
result := pas.Sync[int](Multiply, p1, 3)
```

### Examples

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

**Methods:**

- `Get() T`: Returns the computed value, blocking until it is ready.

### `New`

Creates a new Promise with an optional initial value. The Promise is immediately ready.

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

### `MakeChan`

Creates a channel of `*Promise[T]` with an optional buffer size.

```go
func MakeChan[T any](buffer ...int) chan *Promise[T]
```

**Parameters:**

- `buffer` (optional): Buffer size of the channel.

**Returns:**

- `chan *Promise[T]`: A channel for Promises.

**Usage:**

```go
promiseChan := pas.MakeChan[int](bufferSize)
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
