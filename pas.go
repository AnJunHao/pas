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

// resolve sets the value of the Promise and marks it as ready.
// It can only be called once; subsequent calls will have no effect.
func (p *Promise[T]) resolve(value T) {
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

// newPending creates a pointer to a new Promise holding a value of type T that is not yet ready.
func newPending[T any]() *Promise[T] {
	return &Promise[T]{ready: make(chan struct{})}
}

// Async starts a parallel computation by invoking function f with the provided arguments.
// If any argument is a Promise, it waits for it to be ready before executing f.
// It enforces that function f has exactly one return value of type T.
// It accepts an optional boolean flag as the last argument to enable recursive resolving.
func Async[T any](f interface{}, args ...interface{}) *Promise[T] {
	var recursive bool

	// Detect if the last argument is a boolean flag for recursive resolving
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic(fmt.Sprintf("Async: expected a function, but got %T", f))
	}
	ft := fv.Type()
	numRequiredArgs := ft.NumIn()

	if len(args) == numRequiredArgs+1 {
		if flag, ok := args[len(args)-1].(bool); ok {
			recursive = flag
			args = args[:len(args)-1] // Remove the flag from args
		}
	}

	if len(args) != numRequiredArgs {
		panic(fmt.Sprintf("Async: function expects %d arguments, but got %d", numRequiredArgs, len(args)))
	}

	p := newPending[T]()

	// Start a goroutine to execute the function in parallel
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panics gracefully; in production, consider logging or handling differently
				fmt.Printf("Async: function execution panicked: %v\n", r)
			}
		}()
		// Execute the function and get the result
		output := executeFunction[T](f, recursive, args...)
		// Assign the result to the Promise and signal readiness
		p.resolve(output)
	}()

	return p
}

// Sync executes function f synchronously with the provided arguments.
// If any argument is a Promise, it waits for it to be ready before executing f.
// It enforces that function f has exactly one return value of type T.
// It accepts an optional boolean flag as the last argument to enable recursive resolving.
func Sync[T any](f interface{}, args ...interface{}) T {
	var recursive bool

	// Detect if the last argument is a boolean flag for recursive resolving
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic(fmt.Sprintf("Sync: expected a function, but got %T", f))
	}
	ft := fv.Type()
	numRequiredArgs := ft.NumIn()

	if len(args) == numRequiredArgs+1 {
		if flag, ok := args[len(args)-1].(bool); ok {
			recursive = flag
			args = args[:len(args)-1] // Remove the flag from args
		}
	}

	if len(args) != numRequiredArgs {
		panic(fmt.Sprintf("Sync: function expects %d arguments, but got %d", numRequiredArgs, len(args)))
	}

	// Execute the function and return the result
	return executeFunction[T](f, recursive, args...)
}

// executeFunction is a helper that encapsulates the common logic for Async and Sync.
// It validates the function, resolves arguments based on the expected parameter types,
// invokes the function, and asserts the return type.
// The 'recursive' flag determines whether to resolve promises recursively.
func executeFunction[T any](f interface{}, recursive bool, args ...interface{}) T {
	fv := reflect.ValueOf(f)
	ft := fv.Type()

	// Validate that f is a function
	if fv.Kind() != reflect.Func {
		panic(fmt.Sprintf("pas.executeFunction: expected a function, but got %T", f))
	}

	// Enforce that f has exactly one return value
	if ft.NumOut() != 1 {
		panic(fmt.Sprintf("pas.executeFunction: function must have exactly one return value, but got %d values", ft.NumOut()))
	}

	// Enforce that the number of arguments matches
	if ft.NumIn() != len(args) {
		panic(fmt.Sprintf("pas.executeFunction: function expects %d arguments, but got %d", ft.NumIn(), len(args)))
	}

	// Resolve arguments based on the expected parameter types and the 'recursive' flag
	resolvedArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		expectedType := ft.In(i)
		var resolved interface{}
		var err error

		if recursive {
			// Recursive resolving using resolveValue
			resolved, err = resolveValue(arg, expectedType)
		} else {
			// Shallow resolving: only resolve top-level promises
			resolved, err = shallowResolve(arg, expectedType)
		}

		if err != nil {
			panic(fmt.Sprintf("pas.executeFunction: error resolving argument %d: %v", i, err))
		}

		// Handle nil inputs by setting zero value if necessary
		if resolved == nil {
			resolvedArgs[i] = reflect.Zero(expectedType)
		} else {
			resolvedVal := reflect.ValueOf(resolved)
			// Ensure the resolved argument can be assigned to the expected type
			if !resolvedVal.Type().AssignableTo(expectedType) {
				// Attempt to convert if possible
				if resolvedVal.Type().ConvertibleTo(expectedType) {
					resolvedVal = resolvedVal.Convert(expectedType)
				} else {
					panic(fmt.Sprintf("pas.executeFunction: argument %d has type %s, expected %s",
						i, resolvedVal.Type(), expectedType))
				}
			}
			resolvedArgs[i] = resolvedVal
		}
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

// shallowResolve resolves only the top-level promises without delving into nested structures.
// It returns the resolved value or the original value if it's not a promise.
func shallowResolve(input interface{}, expectedType reflect.Type) (interface{}, error) {
	if input == nil {
		// Return zero value of expectedType
		return reflect.Zero(expectedType).Interface(), nil
	}

	// Handle Promise
	if promise, ok := input.(promiseTypeContract); ok {
		resolved := promise.get()
		return resolved, nil
	}

	// If not a Promise, return as-is
	return input, nil
}

// resolveValue recursively resolves Promises within the input based on the expectedType.
// It handles Promises, pointers, slices, arrays, maps, and nested combinations thereof.
// expectedType defines the type that the resolved value should conform to.
func resolveValue(input interface{}, expectedType reflect.Type) (interface{}, error) {
	if input == nil {
		// Return zero value of expectedType
		return reflect.Zero(expectedType).Interface(), nil
	}

	// Handle Promise
	if promise, ok := input.(promiseTypeContract); ok {
		resolved := promise.get()
		return resolveValue(resolved, expectedType)
	}

	currentType := reflect.TypeOf(input)

	// Handle Pointer Types
	if expectedType.Kind() == reflect.Ptr {
		if currentType.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("expected a pointer of type %s, but got %s", expectedType, currentType)
		}
		// Resolve the value the pointer points to
		if reflect.ValueOf(input).IsNil() {
			return reflect.Zero(expectedType).Interface(), nil
		}
		resolvedElem, err := resolveValue(reflect.ValueOf(input).Elem().Interface(), expectedType.Elem())
		if err != nil {
			return nil, err
		}
		// Create a new pointer of the expected type and set its value
		newPtr := reflect.New(expectedType.Elem())
		newPtr.Elem().Set(reflect.ValueOf(resolvedElem))
		return newPtr.Interface(), nil
	}

	switch expectedType.Kind() {
	case reflect.Slice:
		// Handle Slice Types
		inputVal := reflect.ValueOf(input)
		if inputVal.Kind() != reflect.Slice {
			return nil, fmt.Errorf("expected a slice, but got %s", inputVal.Kind())
		}
		newSlice := reflect.MakeSlice(expectedType, inputVal.Len(), inputVal.Len())
		for i := 0; i < inputVal.Len(); i++ {
			resolvedElem, err := resolveValue(inputVal.Index(i).Interface(), expectedType.Elem())
			if err != nil {
				return nil, fmt.Errorf("error resolving slice element at index %d: %v", i, err)
			}
			newSlice.Index(i).Set(reflect.ValueOf(resolvedElem))
		}
		return newSlice.Interface(), nil

	case reflect.Array:
		// Handle Array Types
		inputVal := reflect.ValueOf(input)
		if inputVal.Kind() != reflect.Array {
			return nil, fmt.Errorf("expected an array, but got %s", inputVal.Kind())
		}
		if inputVal.Len() != expectedType.Len() {
			return nil, fmt.Errorf("expected array of length %d, but got %d", expectedType.Len(), inputVal.Len())
		}
		newArray := reflect.New(expectedType).Elem()
		for i := 0; i < inputVal.Len(); i++ {
			resolvedElem, err := resolveValue(inputVal.Index(i).Interface(), expectedType.Elem())
			if err != nil {
				return nil, fmt.Errorf("error resolving array element at index %d: %v", i, err)
			}
			newArray.Index(i).Set(reflect.ValueOf(resolvedElem))
		}
		return newArray.Interface(), nil

	case reflect.Map:
		// Handle Map Types
		inputVal := reflect.ValueOf(input)
		if inputVal.Kind() != reflect.Map {
			return nil, fmt.Errorf("expected a map, but got %s", inputVal.Kind())
		}
		newMap := reflect.MakeMapWithSize(expectedType, inputVal.Len())
		for _, key := range inputVal.MapKeys() {
			// Resolve the key
			resolvedKey, err := resolveValue(key.Interface(), expectedType.Key())
			if err != nil {
				return nil, fmt.Errorf("error resolving map key %v: %v", key.Interface(), err)
			}
			// Resolve the value
			resolvedValue, err := resolveValue(inputVal.MapIndex(key).Interface(), expectedType.Elem())
			if err != nil {
				return nil, fmt.Errorf("error resolving map value for key %v: %v", resolvedKey, err)
			}
			newMap.SetMapIndex(reflect.ValueOf(resolvedKey), reflect.ValueOf(resolvedValue))
		}
		return newMap.Interface(), nil

	case reflect.Interface:
		// If the expected type is interface{}, return the input as-is after resolving any Promises
		return input, nil

	default:
		// Handle Basic Types and Perform Necessary Conversions
		inputVal := reflect.ValueOf(input)
		if inputVal.Type().AssignableTo(expectedType) {
			return input, nil
		}
		if inputVal.Type().ConvertibleTo(expectedType) {
			return inputVal.Convert(expectedType).Interface(), nil
		}
		return nil, fmt.Errorf("cannot assign or convert %s to %s", inputVal.Type(), expectedType)
	}
}

// shallowResolveArgs processes the arguments, waiting for any Promise to be ready and retrieving its value.
// If an argument is not a Promise, it is used as-is.
// This function is kept for reference but is not used directly as per the new implementation.
func shallowResolveArgs(args ...interface{}) []reflect.Value {
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
