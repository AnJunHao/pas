package pas

import (
	"fmt"
	"testing"
	"time"
)

// Sample functions to use with Async and Sync
func Square(n int) int {
	return n * n
}

func Multiply(a, b int) int {
	return a * b
}

func MultiplyReturnPointer(a, b int) *int {
	result := a * b
	return &result
}

func Add(a, b int) int {
	return a + b
}

// SumWithinRange computes the sum of integers from start to end (inclusive).
func SumWithinRange(start int, end int) int {
	sum := 0
	for i := start; i <= end; i++ {
		sum += i
	}
	return sum
}

func SumSlice(arr []int) int {
	sum := 0
	for _, v := range arr {
		sum += v
	}
	return sum
}

func SumMatrix(matrix [][]int) int {
	sum := 0
	for _, row := range matrix {
		for _, val := range row {
			sum += val
		}
	}
	return sum
}

func SumMap(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

// TestPromise tests the Promise constructor and Get method.
func TestPromise(t *testing.T) {
	p := New(10)
	if val := p.Get(); val != 10 {
		t.Errorf("Expected 10, got %v", val)
	}

	pEmpty := New[int]()
	if val := pEmpty.Get(); val != 0 {
		t.Errorf("Expected 0 (zero value), got %v", val)
	}
}

// TestSingleAsync tests the Async function with a single argument.
func TestSingleAsync(t *testing.T) {
	promise := Async[int](Square, 5)
	value := promise.Get()
	expected := 25
	if value != expected {
		t.Errorf("Expected %d, got %d", expected, value)
	}
}

// TestAsync tests the Async function with multiple arguments.
func TestAsync(t *testing.T) {
	p := New(0)
	for i := 1; i <= 3; i++ {
		sq := Async[int](Square, i)
		p = Async[int](Add, p, sq)
	}

	if val := p.Get(); val != 14 { // 14 = 1*1 + 2*2 + 3*3
		t.Errorf("Expected %v, got %v", 14, val)
	}
}

// TestSync tests the Sync function with mixed Async and Sync calls.
func TestSync(t *testing.T) {
	p := 0
	for i := 1; i <= 3; i++ {
		sq := Async[int](Square, i)
		p = Sync[int](Add, p, sq)
	}

	if val := p; val != 14 { // 14 = 1*1 + 2*2 + 3*3
		t.Errorf("Expected %v, got %v", 14, val)
	}
}

// TestSliceOfPromises verifies that []*Promise[int] instances are correctly resolved to []int values.
func TestSliceOfPromises(t *testing.T) {
	n := 100
	arr := MakeSlice[int](n)
	for i := range arr {
		arr[i] = Async[int](Square, i)
	}
	sum := Sync[int](SumSlice, arr)
	expected := 0
	for i := 0; i < n; i++ {
		expected += i * i
	}
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// TestNestedSlicesOfPromises verifies that [][]*Promise[int] instances are correctly resolved to [][]int values.
func TestNestedSlicesOfPromises(t *testing.T) {
	n := 50
	nestedSlice := make([][]*Promise[int], n)
	for i := 0; i < n; i++ {
		inner := MakeSlice[int](n)
		for j := 0; j < n; j++ {
			inner[j] = Async[int](Multiply, i, j)
		}
		nestedSlice[i] = inner
	}
	sum := Sync[int](SumMatrix, nestedSlice)
	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			expected += i * j
		}
	}
	if sum != expected {
		t.Errorf("Expected nested sum %d, got %d", expected, sum)
	}
}

// TestMapOfPromises verifies that map[string]*Promise[int] instances are correctly resolved to map[string]int values.
func TestMapOfPromises(t *testing.T) {
	m := MakeMap[string, int](5)
	keys := []string{"a", "b", "c", "d", "e"}
	for _, key := range keys {
		m[key] = Async[int](Square, len(key)) // Square the length of the key
	}
	sum := Sync[int](SumMap, m)
	expected := 0
	for _, key := range keys {
		expected += len(key) * len(key)
	}
	if sum != expected {
		t.Errorf("Expected map sum %d, got %d", expected, sum)
	}
}

// TestNestedMaps verifies that map[string]map[string]*Promise[int] instances are correctly resolved to map[string]map[string]int values.
func TestNestedMaps(t *testing.T) {
	n := 5
	outerMap := make(map[string]map[string]*Promise[int], n)
	for i := 0; i < n; i++ {
		innerMap := MakeMap[string, int](n)
		for j := 0; j < n; j++ {
			key := fmt.Sprintf("key_%d_%d", i, j)
			innerMap[key] = Async[int](Multiply, i, j)
		}
		outerMap[fmt.Sprintf("outer_%d", i)] = innerMap
	}
	// Define a function to sum all values in a nested map
	sumNested := func(m map[string]map[string]int) int {
		sum := 0
		for _, inner := range m {
			for _, v := range inner {
				sum += v
			}
		}
		return sum
	}
	sum := Sync[int](sumNested, outerMap)
	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			expected += i * j
		}
	}
	if sum != expected {
		t.Errorf("Expected nested map sum %d, got %d", expected, sum)
	}
}

// TestMixedNestedStructures tests the resolution of complex nested structures
// that combine slices and maps, containing both Promises and non-Promises.
func TestMixedNestedStructures(t *testing.T) {
	n := 10
	// Create a map where each key maps to a slice of Promises
	mappedSlices := make(map[string][]*Promise[int], n)
	for i := 0; i < n; i++ {
		promises := MakeSlice[int](n)
		for j := 0; j < n; j++ {
			promises[j] = Async[int](Multiply, i, j)
		}
		mappedSlices[fmt.Sprintf("map_%d", i)] = promises
	}
	// Define a function to sum all values in the map of slices
	sumMixed := func(m map[string][]int) int {
		sum := 0
		for _, slice := range m {
			for _, val := range slice {
				sum += val
			}
		}
		return sum
	}
	sum := Sync[int](sumMixed, mappedSlices)
	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			expected += i * j
		}
	}
	if sum != expected {
		t.Errorf("Expected mixed nested sum %d, got %d", expected, sum)
	}
}

// TestEmptySlice verifies that the Sync function correctly handles empty slices.
func TestEmptySlice(t *testing.T) {
	emptySlice := MakeSlice[int](0)
	sum := Sync[int](SumSlice, emptySlice)
	expected := 0
	if sum != expected {
		t.Errorf("Expected sum %d for empty slice, got %d", expected, sum)
	}
}

// TestEmptyMap verifies that the Sync function correctly handles empty maps.
func TestEmptyMap(t *testing.T) {
	emptyMap := make(map[string]int)
	sum := Sync[int](SumMap, emptyMap)
	expected := 0
	if sum != expected {
		t.Errorf("Expected sum %d for empty map, got %d", expected, sum)
	}
}

// TestNilInput verifies that a nil input is correctly handled
// and resolved to the zero value of the expected type.
func TestNilInput(t *testing.T) {
	var nilSlice []*Promise[int] = nil
	sum := Sync[int](SumSlice, nilSlice)
	expected := 0
	if sum != expected {
		t.Errorf("Expected sum %d for nil slice, got %d", expected, sum)
	}

	var nilMap map[string]*Promise[int] = nil
	sumMap := Sync[int](SumMap, nilMap)
	if sumMap != 0 {
		t.Errorf("Expected sum %d for nil map, got %d", 0, sumMap)
	}
}

// TestMixedPromisesInSlice verifies that a slice containing
// both *Promise[int] and regular int values is correctly resolved,
// with Promises being resolved and non-Promises being used as-is.
func TestMixedPromisesInSlice(t *testing.T) {
	n := 10
	mixedSlice := make([]interface{}, n)
	expectedSum := 0
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			promise := Async[int](Square, i)
			mixedSlice[i] = promise
			expectedSum += i * i
		} else {
			value := i
			mixedSlice[i] = value
			expectedSum += i
		}
	}
	// Define a function to sum a slice of ints
	sumFunc := func(arr []int) int {
		sum := 0
		for _, v := range arr {
			sum += v
		}
		return sum
	}
	sum := Sync[int](sumFunc, mixedSlice)
	if sum != expectedSum {
		t.Errorf("Expected mixed sum %d, got %d", expectedSum, sum)
	}
}

// TestMixedPromisesInMap verifies that a map containing
// both *Promise[int] and regular int values is correctly resolved,
// with Promises being resolved and non-Promises being used as-is.
func TestMixedPromisesInMap(t *testing.T) {
	mixedMap := make(map[string]interface{})
	expectedSum := 0
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		if i%2 == 0 {
			promise := Async[int](Multiply, i, i)
			mixedMap[key] = promise
			expectedSum += i * i
		} else {
			value := i
			mixedMap[key] = value
			expectedSum += i
		}
	}
	// Define a function to sum a map of ints
	sumFunc := func(m map[string]int) int {
		sum := 0
		for _, v := range m {
			sum += v
		}
		return sum
	}
	sum := Sync[int](sumFunc, mixedMap)
	if sum != expectedSum {
		t.Errorf("Expected mixed map sum %d, got %d", expectedSum, sum)
	}
}

// TestDeeplyNestedStructures tests the resolution of
// highly nested structures combining slices and maps at multiple levels.
func TestDeeplyNestedStructures(t *testing.T) {
	n := 5
	deeplyNested := MakeSlice[string](n) // Outer slice: []*Promise[string]
	for i := 0; i < n; i++ {
		innerMap := MakeMap[string, int](n)
		for j := 0; j < n; j++ {
			key := fmt.Sprintf("key_%d_%d", i, j)
			innerMap[key] = Async[int](Multiply, i+1, j+1) // Avoiding zero multiplications
		}
		deeplyNested[i] = Async[string](func(m map[string]int) string {
			sum := 0
			for _, v := range m {
				sum += v
			}
			return fmt.Sprintf("Sum: %d", sum)
		}, innerMap)
	}
	// Define a function to concatenate strings from the slice
	concatFunc := func(arr []string) string {
		result := ""
		for _, s := range arr {
			result += s + ";"
		}
		return result
	}
	concat := Sync[string](concatFunc, deeplyNested)
	// Calculate expected sum
	expectedConcat := ""
	for i := 0; i < n; i++ {
		sum := 0
		for j := 0; j < n; j++ {
			sum += (i + 1) * (j + 1)
		}
		expectedConcat += fmt.Sprintf("Sum: %d;", sum)
	}
	if concat != expectedConcat {
		t.Errorf("Expected concatenated string '%s', got '%s'", expectedConcat, concat)
	}
}

// TestPromisesWithDifferentTypes tests that Promises holding
// different types are correctly resolved and type-safe within a heterogeneous structure.
func TestPromisesWithDifferentTypes(t *testing.T) {
	n := 5
	mixedSlice := make([]interface{}, n)
	expectedConcat := ""
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			// Promises of int
			promise := Async[int](Multiply, i, i)
			mixedSlice[i] = promise
		} else {
			// Promises of string
			msg := fmt.Sprintf("Number %d squared is %d", i, i*i)
			promise := Async[string](func(s string) string {
				return s
			}, msg)
			mixedSlice[i] = promise
			expectedConcat += msg
		}
	}
	// Define a function to concatenate strings and sum ints
	type resultStruct struct {
		sum    int
		concat string
	}
	processMixedStruct := func(arr []interface{}) resultStruct {
		sum := 0
		concat := ""
		for _, item := range arr {
			switch v := item.(type) {
			case int:
				sum += v
			case string:
				concat += v
			default:
			}
		}
		return resultStruct{sum: sum, concat: concat}
	}
	sumConcat := Sync[resultStruct](processMixedStruct, mixedSlice)
	expectedSum := 0
	for i := 0; i < n; i += 2 {
		expectedSum += i * i
	}
	if sumConcat.sum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, sumConcat.sum)
	}
	if sumConcat.concat != expectedConcat {
		t.Errorf("Expected concat '%s', got '%s'", expectedConcat, sumConcat.concat)
	}
}

// TestPromisesWithinPointers tests resolving promises that return pointers within slices.
func TestPromisesWithinPointers(t *testing.T) {
	n := 10
	ptrSlice := MakeSlice[*int](n, n)
	for i := 0; i < n; i++ {
		// MultiplyReturnPointer returns a pointer to an int
		promise := Async[*int](MultiplyReturnPointer, i+1, 2) // i+1 to avoid zero
		ptrSlice[i] = promise
	}

	// Define a function to dereference pointers and sum the ints
	sumDeref := func(arr []*int) int {
		sum := 0
		for _, ptr := range arr {
			if ptr != nil {
				sum += *ptr
			}
		}
		return sum
	}

	sum := Sync[int](sumDeref, ptrSlice)

	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		expected += (i + 1) * 2
	}

	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

func TestNestedStructuresWithZeroValues(t *testing.T) {
	// Create a map where some Promises resolve to zero
	m := MakeMap[string, int]()
	m["a"] = Async[int](Square, 0)      // Resolves to 0
	m["b"] = Async[int](Square, 2)      // Resolves to 4
	m["c"] = Async[int](Multiply, 0, 5) // Resolves to 0
	m["d"] = Async[int](Multiply, 3, 3) // Resolves to 9

	sum := Sync[int](SumMap, m)
	expected := 0 + 4 + 0 + 9 // Sum is 13
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

type Person struct {
	Name string
	Age  int
}

func CreatePerson(name string, age int) Person {
	return Person{Name: name, Age: age}
}

func SumAges(people []Person) int {
	sum := 0
	for _, p := range people {
		sum += p.Age
	}
	return sum
}

func TestPromisesWithComplexTypes(t *testing.T) {
	n := 5
	peoplePromises := make([]*Promise[Person], n)
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}
	ages := []int{30, 25, 35, 28, 22}
	for i := 0; i < n; i++ {
		peoplePromises[i] = Async[Person](CreatePerson, names[i], ages[i])
	}
	sumAges := Sync[int](SumAges, peoplePromises)
	expected := 30 + 25 + 35 + 28 + 22 // Sum is 140
	if sumAges != expected {
		t.Errorf("Expected sum of ages %d, got %d", expected, sumAges)
	}
}

type Employee struct {
	ID     int
	Salary *Promise[int] // It is deprecated to include Promise in structs. Here for testing purposes.
}

func SumSalaries(employees []Employee) int {
	sum := 0
	for _, e := range employees {
		sum += e.Salary.Get()
	}
	return sum
}

func TestPromisesWithinStructs(t *testing.T) {
	n := 5
	employees := make([]Employee, n)
	expectedSum := 0
	for i := 0; i < n; i++ {
		employees[i].ID = i + 1
		employees[i].Salary = Async[int](Multiply, (i+1)*1000, 1) // Salaries: 1000, 2000, ..., 5000
		expectedSum += (i + 1) * 1000
	}
	sum := Sync[int](SumSalaries, employees)
	if sum != expectedSum {
		t.Errorf("Expected sum of salaries %d, got %d", expectedSum, sum)
	}
}

// ConcatStrings concatenates all string elements in a slice.
func ConcatStrings(arr []string) string {
	result := ""
	for _, s := range arr {
		result += s
	}
	return result
}

// TestInterfaceSlice ensures that a slice of interface{}
// containing both promises and native types (e.g., int, string) is correctly resolved.
func TestInterfaceSlice(t *testing.T) {
	n := 5
	mixedInterfaceSlice := make([]interface{}, n)
	expectedSum := 0
	expectedConcat := ""

	for i := 0; i < n; i++ {
		if i%2 == 0 {
			// Even indices: Promises of int
			promise := Async[int](Multiply, i+1, 3) // (i+1)*3
			mixedInterfaceSlice[i] = promise
			expectedSum += (i + 1) * 3
		} else {
			// Odd indices: Promises of string
			msg := fmt.Sprintf("msg%d", i)
			promise := Async[string](func(s string) string {
				return s + "_resolved"
			}, msg)
			mixedInterfaceSlice[i] = promise
			expectedConcat += msg + "_resolved"
		}
	}

	// Define a function to process mixed interface{} slice
	processMixedInterfaceSlice := func(arr []interface{}) struct {
		sum    int
		concat string
	} {
		sum := 0
		concat := ""
		for _, item := range arr {
			switch v := item.(type) {
			case int:
				sum += v
			case string:
				concat += v
			default:
				// Handle unexpected types if necessary
			}
		}
		return struct {
			sum    int
			concat string
		}{sum: sum, concat: concat}
	}

	// Execute Sync
	result := Sync[struct {
		sum    int
		concat string
	}](processMixedInterfaceSlice, mixedInterfaceSlice)

	// Assertions
	if result.sum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, result.sum)
	}
	if result.concat != expectedConcat {
		t.Errorf("Expected concat '%s', got '%s'", expectedConcat, result.concat)
	}
}

// ConcatenateMapStrings concatenates all string values in a map.
func ConcatenateMapStrings(m map[string]string) string {
	result := ""
	for _, s := range m {
		result += s
	}
	return result
}

// SumMapInts sums all integer values in a map.
func SumMapInts(m map[string]int) int {
	sum := 0
	for _, n := range m {
		sum += n
	}
	return sum
}

// TestInterfaceMap verifies that a map with values of type interface{},
// containing promises of different types (int, string), is correctly resolved.
func TestInterfaceMap(t *testing.T) {
	n := 5
	mixedInterfaceMap := make(map[string]interface{}, n)
	expectedSum := 0
	expectedConcat := ""

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		if i%2 == 0 {
			// Even keys: Promises of int
			promise := Async[int](Multiply, i+2, 4) // (i+2)*4
			mixedInterfaceMap[key] = promise
			expectedSum += (i + 2) * 4
		} else {
			// Odd keys: Promises of string
			msg := fmt.Sprintf("value%d", 7777777)
			promise := Async[string](func(s string) string {
				return s + "_computed"
			}, msg)
			mixedInterfaceMap[key] = promise
			expectedConcat += msg + "_computed"
		}
	}

	// Define a function to process mixed interface{} map
	processMixedInterfaceMap := func(m map[string]interface{}) struct {
		sum    int
		concat string
	} {
		sum := 0
		concat := ""
		for _, v := range m {
			switch val := v.(type) {
			case int:
				sum += val
			case string:
				concat += val
			default:
				// Handle unexpected types if necessary
			}
		}
		return struct {
			sum    int
			concat string
		}{sum: sum, concat: concat}
	}

	// Execute Sync
	result := Sync[struct {
		sum    int
		concat string
	}](processMixedInterfaceMap, mixedInterfaceMap)

	// Assertions
	if result.sum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, result.sum)
	}
	if result.concat != expectedConcat {
		t.Errorf("Expected concat '%s', got '%s'", expectedConcat, result.concat)
	}
}

// SumPointersSlice sums the dereferenced integers from a slice of *int.
func SumPointersSlice(arr []*int) int {
	sum := 0
	for _, ptr := range arr {
		if ptr != nil {
			sum += *ptr
		}
	}
	return sum
}

// SumPointersMap sums the dereferenced integers from a map of string to *int.
func SumPointersMap(m map[string]*int) int {
	sum := 0
	for _, ptr := range m {
		if ptr != nil {
			sum += *ptr
		}
	}
	return sum
}

// TestPointersInSliceOfPromises tests the resolution of a slice containing pointers to Promises that hold pointers to ints.
func TestPointersInSliceOfPromises(t *testing.T) {
	n := 10
	// Create a slice of *Promise[*int]
	ptrPromiseSlice := MakeSlice[*int](n, n)
	for i := 0; i < n; i++ {
		// Each Promise resolves to a pointer to int
		promise := Async[*int](MultiplyReturnPointer, i+1, 3) // Multiply (i+1) by 3
		ptrPromiseSlice[i] = promise
	}

	// Define a function to sum dereferenced *int values from a slice
	sumPointersSlice := func(arr []*int) int {
		return SumPointersSlice(arr)
	}

	// Execute Sync to resolve all Promises and compute the sum
	sum := Sync[int](sumPointersSlice, ptrPromiseSlice)

	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		expected += (i + 1) * 3
	}

	// Assertion
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// TestPointersInMapOfPromises tests the resolution of a map containing pointers to Promises that hold pointers to ints.
func TestPointersInMapOfPromises(t *testing.T) {
	n := 10
	// Create a map of string to *Promise[*int]
	ptrPromiseMap := MakeMap[string, *int](n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		// Each Promise resolves to a pointer to int
		promise := Async[*int](MultiplyReturnPointer, (i+1)*2, 4) // Multiply (i+1)*2 by 4
		ptrPromiseMap[key] = promise
	}

	// Define a function to sum dereferenced *int values from a map
	sumPointersMapFunc := func(m map[string]*int) int {
		return SumPointersMap(m)
	}

	// Execute Sync to resolve all Promises and compute the sum
	sum := Sync[int](sumPointersMapFunc, ptrPromiseMap)

	// Calculate expected sum
	expected := 0
	for i := 0; i < n; i++ {
		expected += (i + 1) * 2 * 4
	}

	// Assertion
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// Inner represents a simple struct containing a pointer to an int.
type Inner struct {
	Value *int
}

// Outer represents a composite struct containing a pointer to Inner.
type Outer struct {
	InnerPtr *Inner
}

// CreateOuter initializes an Outer struct with nested pointers.
func CreateOuter(a int) *Outer {
	return &Outer{
		InnerPtr: &Inner{
			Value: &a,
		},
	}
}

// TestPromisesWithinComplexPointers_Slice tests resolving a slice containing pointers to Promises,
// each of which resolves to a pointer to an Outer struct containing a nested pointer.
func TestPromisesWithinComplexPointers_Slice(t *testing.T) {
	n := 10
	// Create a slice of *Promise[*Outer]
	outerPromiseSlice := MakeSlice[*Outer](n, n)

	// Populate the slice with Promises that resolve to *Outer
	for i := 0; i < n; i++ {
		// Each Promise resolves to an Outer containing an Inner with a pointer to (i+1)*5
		promise := Async[*Outer](CreateOuter, (i+1)*5)
		outerPromiseSlice[i] = promise
	}

	// Define a function to sum the dereferenced values from a slice of *Outer
	sumOuterSlice := func(arr []*Outer) int {
		sum := 0
		for _, outer := range arr {
			if outer != nil && outer.InnerPtr != nil && outer.InnerPtr.Value != nil {
				sum += *outer.InnerPtr.Value
			}
		}
		return sum
	}

	// Execute Sync to resolve all Promises and compute the sum
	sum := Sync[int](sumOuterSlice, outerPromiseSlice)

	// Calculate expected sum
	expected := 0
	for i := 1; i <= n; i++ {
		expected += i * 5
	}

	// Assertion
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// TestPromisesWithinComplexPointers_Map tests resolving a map containing pointers to Promises,
// each of which resolves to a pointer to an Outer struct containing a nested pointer.
func TestPromisesWithinComplexPointers_Map(t *testing.T) {
	n := 10
	// Create a map of string to *Promise[*Outer]
	outerPromiseMap := MakeMap[string, *Outer](n)

	// Populate the map with Promises that resolve to *Outer
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key%d", i)
		// Each Promise resolves to an Outer containing an Inner with a pointer to (i+1)*7
		promise := Async[*Outer](CreateOuter, (i+1)*7)
		outerPromiseMap[key] = promise
	}

	// Define a function to sum the dereferenced values from a map of *Outer
	sumOuterMap := func(m map[string]*Outer) int {
		sum := 0
		for _, outer := range m {
			if outer != nil && outer.InnerPtr != nil && outer.InnerPtr.Value != nil {
				sum += *outer.InnerPtr.Value
			}
		}
		return sum
	}

	// Execute Sync to resolve all Promises and compute the sum
	sum := Sync[int](sumOuterMap, outerPromiseMap)

	// Calculate expected sum
	expected := 0
	for i := 1; i <= n; i++ {
		expected += i * 7
	}

	// Assertion
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// SumDeepNestedInts sums all integer values in a deeply nested structure:
// pointer to slice of pointers to map[int]*int
func SumDeepNestedInts(ppsm *[]*map[int]*int) int {
	sum := 0
	for _, pmPtr := range *ppsm {
		for _, v := range *pmPtr {
			if v != nil {
				sum += *v
			}
		}
	}
	return sum
}

// TestDeeplyNestedPointerSliceMap tests resolving a pointer to slice of pointers to map[int]*int,
// where the map values are promises that resolve to *int.
func TestDeeplyNestedPointerSliceMap(t *testing.T) {
	// Define the size of the nested structures
	numSlices := 3
	numEntriesPerMap := 2

	// Create a slice of pointers to maps
	// This will be the input to the SumDeepNestedInts function
	// Note that the function input expects: *[]*map[int]*int
	// The innermost *int will be a Promise that resolves to an int
	sliceOfMaps := make([]*map[int]*Promise[*int], numSlices)
	for i := 0; i < numSlices; i++ {
		// For each slice element, create a map[int]*Promise[*int]
		promiseMap := make(map[int]*Promise[*int], numEntriesPerMap)
		for j := 0; j < numEntriesPerMap; j++ {
			key := i*numEntriesPerMap + j
			// Each map value is a Promise that resolves to *int
			val := Async[*int](MultiplyReturnPointer, key, 10) // val = key * 10
			promiseMap[key] = val
		}
		// Assign the promise map to the slice
		sliceOfMaps[i] = &promiseMap
	}

	// Create a pointer to the slice
	pointerToSlice := &sliceOfMaps

	// Execute Sync with the SumDeepNestedInts function
	sum := Sync[int](SumDeepNestedInts, pointerToSlice)

	// Calculate the expected sum
	expected := 0
	for i := 0; i < numSlices; i++ {
		for j := 0; j < numEntriesPerMap; j++ {
			expected += (i*numEntriesPerMap + j) * 10
		}
	}

	// Assertion
	if sum != expected {
		t.Errorf("Expected sum %d, got %d", expected, sum)
	}
}

// TransformMixedStructures transforms a complex nested structure by performing operations on its elements.
// It takes a pointer to a slice of maps containing arrays of pointers to strings.
func TransformMixedStructures(psm *[]map[string][2]*string) map[string][2]string {
	transformed := make(map[string][2]string)
	for _, m := range *psm {
		for key, arrayPtr := range m {
			var newArray [2]string
			for i, strPtr := range arrayPtr {
				if strPtr != nil {
					newArray[i] = *strPtr
				} else {
					newArray[i] = ""
				}
			}
			transformed[key] = newArray
		}
	}
	return transformed
}

func ConcatTwoStringsReturnPointer(str1, str2 string) *string {
	concat := str1 + str2
	return &concat
}

// TestMixedNestedPointerSliceMap tests resolving a complex nested structure with mixed types and Promises.
func TestMixedNestedPointerSliceMap(t *testing.T) {
	// Define the size of the nested structures
	numMaps := 2
	numEntriesPerMap := 2
	arraySize := 2

	// Create a slice of maps
	sliceOfMaps := make([]map[string][2]*Promise[*string], numMaps)
	for i := 0; i < numMaps; i++ {
		currentMap := make(map[string][2]*Promise[*string], numEntriesPerMap)
		for j := 0; j < numEntriesPerMap; j++ {
			// Use fmt.Sprintf to construct the key properly
			key := fmt.Sprintf("Key_%c%d", 'A'+i, j+1)

			var arrayOfPromises [2]*Promise[*string]
			for k := 0; k < arraySize; k++ {
				str1 := "Hello_"
				// Use fmt.Sprintf to ensure proper string construction
				str2 := fmt.Sprintf("%c", 'a'+rune(i*2+j))

				promise := Async[*string](ConcatTwoStringsReturnPointer, str1, str2) // e.g., "Hello_a", "Hello_b", etc.
				arrayOfPromises[k] = promise
			}
			currentMap[key] = arrayOfPromises
		}
		sliceOfMaps[i] = currentMap
	}

	// Create a pointer to the slice
	pointerToSlice := &sliceOfMaps

	// Execute Sync with the TransformMixedStructures function
	transformed := Sync[map[string][2]string](TransformMixedStructures, pointerToSlice)

	// Calculate the expected transformed map
	expected := make(map[string][2]string)
	for i := 0; i < numMaps; i++ {
		for j := 0; j < numEntriesPerMap; j++ {
			// Use the same key construction method
			key := fmt.Sprintf("Key_%c%d", 'A'+i, j+1)

			var arr [2]string
			for k := 0; k < arraySize; k++ {
				concatStr := fmt.Sprintf("Hello_%c", 'a'+rune(i*2+j))
				arr[k] = concatStr
			}
			expected[key] = arr
		}
	}

	// Assertions
	if len(transformed) != len(expected) {
		t.Fatalf("TestMixedNestedPromises: Expected transformed map length %d, got %d", len(expected), len(transformed))
	}
	for key, expectedArr := range expected {
		transformedArr, exists := transformed[key]
		if !exists {
			t.Errorf("TestMixedNestedPromises: Key %s missing in transformed map", key)
			continue
		}
		for i := 0; i < arraySize; i++ {
			if transformedArr[i] != expectedArr[i] {
				t.Errorf("TestMixedNestedPromises: For key %s, index %d: expected %s, got %s", key, i, expectedArr[i], transformedArr[i])
			}
		}
	}

	t.Logf("TestMixedNestedPromises passed: transformed map matches expected values")
}

// TestParallelSum tests the parallel sum implementation against the sequential sum.
func TestParallelSum(t *testing.T) {
	// Adjust n for faster test execution
	n := 1000000000
	numWorkers := 20

	startTime := time.Now()

	// Parallel Sum
	parSum := New(0)
	for i := 0; i < numWorkers; i++ {
		// Define the start and end for each worker
		start := i*n/numWorkers + 1
		end := (i + 1) * n / numWorkers

		// Start an asynchronous computation for the sum within the range
		s := Async[int](SumWithinRange, start, end)

		// Aggregate the results by adding them asynchronously
		parSum = Async[int](Add, parSum, s)
	}

	// Retrieve the parallel sum result
	parallelResult := parSum.Get()
	parallelDuration := time.Since(startTime)

	// Log the parallel computation result and duration
	t.Logf("Parallel Sum Result: %d", parallelResult)
	t.Logf("Parallel Sum took: %v", parallelDuration)

	// Sequential Sum
	startTime = time.Now()
	seqSum := SumWithinRange(1, n)
	seqDuration := time.Since(startTime)

	// Log the sequential computation result and duration
	t.Logf("Sequential Sum Result: %d", seqSum)
	t.Logf("Sequential Sum took: %v", seqDuration)

	// Validate that both sums are equal
	if seqSum != parallelResult {
		t.Errorf("Mismatch in sums: Sequential Sum = %d, Parallel Sum = %d", seqSum, parallelResult)
	} else {
		t.Log("Success: Sequential and Parallel results match.")
	}

	// Optional: Compare performance (not typically done in unit tests)

	if parallelDuration >= seqDuration {
		t.Errorf("Parallel execution took longer or equal time compared to sequential execution. Parallel: %v, Sequential: %v", parallelDuration, seqDuration)
	} else {
		t.Log("Parallel execution is faster than sequential execution.")
	}
}

// TestParallelSumWithSliceOfPromises verifies that the parallel sum implementation with a slice of promises is correct.
func TestParallelSumWithSliceOfPromises(t *testing.T) {
	n := 1000000000
	numWorkers := 20

	// Parallel Execution
	startTime := time.Now()
	arr := MakeSlice[int](numWorkers)
	batchSize := n / numWorkers
	for i := range arr {
		arr[i] = Async[int](SumWithinRange, i*batchSize+1, (i+1)*batchSize)
	}
	sum := Sync[int](SumSlice, arr)
	parallelDuration := time.Since(startTime)

	// Log the parallel computation result and duration
	t.Logf("Parallel Sum Result: %d", sum)
	t.Logf("Parallel Sum took: %v", parallelDuration)

	// Sequential Execution
	startTime = time.Now()
	arrSeq := make([]int, numWorkers)
	for i := range arrSeq {
		arrSeq[i] = SumWithinRange(i*batchSize+1, (i+1)*batchSize)
	}
	seqSum := SumSlice(arrSeq)
	seqDuration := time.Since(startTime)

	// Log the sequential computation result and duration
	t.Logf("Sequential Sum Result: %d", seqSum)
	t.Logf("Sequential Sum took: %v", seqDuration)
}
