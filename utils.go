package pas

// MakeSlice creates a slice of *Promise[T] with the specified length and capacity.
// Usage example: promises := MakeSlice[int](5)
// This is equivalent to:
// promises := make([]*Promise[int], 5)
// for i := range promises {
// 	promises[i] = New[int]()
// }
func MakeSlice[T any](length int, capacity ...int) []*Promise[T] {
	capVal := length
	if len(capacity) > 0 {
		capVal = capacity[0]
	}
	slice := make([]*Promise[T], length, capVal)
	for i := range slice {
		slice[i] = New[T]()
	}
	return slice
}

// MakeMap creates a map with keys of type K and values of type *Promise[V].
// The optional size parameter can be used to hint the initial size of the map.
// Usage example: promiseMap := MakeMap[string, int](10)
// This is equivalent to:
// promiseMap := make(map[string]*Promise[int], 10)
func MakeMap[K comparable, V any](size ...int) map[K]*Promise[V] {
	var m map[K]*Promise[V]
	if len(size) > 0 {
		m = make(map[K]*Promise[V], size[0])
	} else {
		m = make(map[K]*Promise[V])
	}
	return m
}
