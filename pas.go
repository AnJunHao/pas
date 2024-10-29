package pas

import (
	"fmt"
	"reflect"
	"sync"
)

// promiseTypeContract is an internal interface that identifies a Promise.
// It has an unexported method to prevent external packages from implementing it.
type promiseTypeContract interface { // unexported
	get() interface{}
}

// Promise represents a parallel variable holding a value of type T.
type Promise[T any] struct {
	value T
	ready chan struct{}
	once  sync.Once
}

// Get returns the computed value, blocking until it is ready.
func (p *Promise[T]) Get() T {
	<-p.ready
	return p.value
}

// Resolve sets the value of the Promise and marks it as ready.
// It can only be called once; subsequent calls will have no effect.
func (p *Promise[T]) Resolve(value T) {
	p.once.Do(func() {
		p.value = value
		close(p.ready)
	})
}

// get is an unexported method to satisfy the promiseTypeContract interface.
// It retrieves the value held by the promise, blocking until it's ready.
func (p *Promise[T]) get() interface{} {
	<-p.ready
	return p.value
}

// New creates a pointer to a new Promise holding a value of type T.
func New[T any](values ...T) *Promise[T] {
	p := &Promise[T]{ready: make(chan struct{})}
	if len(values) == 0 {
		// Do not set p.value; leave it zero-valued
	} else if len(values) == 1 {
		p.value = values[0]
	} else {
		panic(fmt.Sprintf("New: expected at most one value, got %d values", len(values)))
	}
	p.once.Do(func() {
		close(p.ready)
	})
	return p
}

// NewPending creates a pointer to a new Promise holding a value of type T that is not yet ready.
func NewPending[T any]() *Promise[T] {
	return &Promise[T]{ready: make(chan struct{})}
}

// MakeSlice creates a slice of *Promise[T] with the specified length and capacity.
// Usage example: promises := MakeSlice[int](5)
func MakeSlice[T any](length int, capacity ...int) []*Promise[T] {
	cap := length
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	slice := make([]*Promise[T], length, cap)
	for i := range slice {
		slice[i] = New[T]()
	}
	return slice
}

// MakePendingSlice creates a slice of *Promise[T] with the specified length and capacity.
// Usage example: promises := MakePendingSlice[int](5)
func MakePendingSlice[T any](length int, capacity ...int) []*Promise[T] {
	cap := length
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	slice := make([]*Promise[T], length, cap)
	for i := range slice {
		slice[i] = NewPending[T]()
	}
	return slice
}

// MakeMap creates a map with keys of type K and values of type *Promise[V].
// The optional size parameter can be used to hint the initial size of the map.
// Usage example: promiseMap := MakeMap[string, int](10)
func MakeMap[K comparable, V any](size ...int) map[K]*Promise[V] {
	var m map[K]*Promise[V]
	if len(size) > 0 {
		m = make(map[K]*Promise[V], size[0])
	} else {
		m = make(map[K]*Promise[V])
	}
	return m
}

// MakeChan creates a channel of *Promise[T] with an optional buffer size.
// Usage example: promiseChan := MakeChan[int](bufferSize)
func MakeChan[T any](buffer ...int) chan *Promise[T] {
	if len(buffer) > 0 {
		return make(chan *Promise[T], buffer[0])
	}
	return make(chan *Promise[T])
}

// Async starts a parallel computation by invoking function f with the provided arguments.
// If any argument is a Promise, it waits for it to be ready before executing f.
// It enforces that function f has exactly one return value of type T.
func Async[T any](f interface{}, args ...interface{}) *Promise[T] {
	p := NewPending[T]()

	// Start a goroutine to execute the function in parallel
	go func() {
		// Execute the function and get the result
		output := executeFunction[T](f, args...)
		// Assign the result to the Promise and signal readiness
		p.Resolve(output)
	}()

	return p
}

// Sync executes function f synchronously with the provided arguments.
// If any argument is a Promise, it waits for it to be ready before executing f.
// It enforces that function f has exactly one return value of type T.
func Sync[T any](f interface{}, args ...interface{}) T {
	// Execute the function and return the result
	return executeFunction[T](f, args...)
}

// executeFunction is a helper that encapsulates the common logic for Async and Sync.
// It validates the function, resolves arguments, invokes the function, and asserts the return type.
func executeFunction[T any](f interface{}, args ...interface{}) T {
	resolvedArgs := resolveArgs(args...)

	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic(fmt.Sprintf("pas.executeFunction: expected a function, but got %T", f))
	}

	// Enforce that f has exactly one return value
	if fv.Type().NumOut() != 1 {
		panic(fmt.Sprintf("pas.executeFunction: function must have exactly one return value, but got %d values", fv.Type().NumOut()))
	}

	// Call the function with the resolved arguments
	results := fv.Call(resolvedArgs)
	if len(results) != 1 {
		panic(fmt.Sprintf("pas.executeFunction: function must return exactly one value, but got %d values", len(results)))
	}

	// Assert that the return type matches T
	output, ok := results[0].Interface().(T)
	if !ok {
		panic(fmt.Sprintf("pas.executeFunction: return type of function does not match generic type. Expected %T, got %T",
			*new(T), results[0].Interface()))
	}

	return output
}

// resolveArgs processes the arguments, waiting for any Promise to be ready and retrieving its value.
// If an argument is not a Promise, it is used as-is.
func resolveArgs(args ...interface{}) []reflect.Value {
	resolved := make([]reflect.Value, len(args))

	for i, arg := range args {
		// Type assertion to check if arg implements promiseTypeContract
		if promiseArg, ok := arg.(promiseTypeContract); ok {
			// Retrieve the value from the promise
			value := promiseArg.get()
			resolved[i] = reflect.ValueOf(value)
		} else {
			// Use the argument as-is
			resolved[i] = reflect.ValueOf(arg)
		}
	}

	return resolved
}
