package pas

import (
	"testing"
	"time"
)

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

func Square(x int) int {
	return x * x
}

func Add(a, b int) int {
	return a + b
}

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

// SumWithinRange computes the sum of integers from start to end (inclusive).
func SumWithinRange(start int, end int) int {
	sum := 0
	for i := start; i <= end; i++ {
		sum += i
	}
	return sum
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
