# PAS

PAS is a Go package that provides a simplistic implementation of promises for concurrent programming. It allows you to create promises, execute asynchronous computations, and synchronize results.

## Simplistic API

- `Promise[T](value)` creates a promise with the given value.
  - `promise.Get()` returns the value of the promise.
- `Async[T](fn, ...args)`: Asynchronously executes the given function with the provided arguments and returns a promise.
- `Sync[T](fn, ...args)`: Synchronously executes the given function with the provided arguments and returns the result.

Both `Async` and `Sync` accepts promises as arguments.

## Installation

```bash
go get github.com/AnJunHao/pas
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.