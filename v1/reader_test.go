package rateLimitedReader

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestRateLimitedReader_BasicRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 4
	data := strings.Repeat("A", dataSize)
	reader := strings.NewReader(data)
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)
	buffer := make([]byte, dataSize/partsAmount)

	var totalRead int
	start := time.Now()

	for totalRead < dataSize {
		read, err := ratelimitedReader.Read(buffer)
		totalRead += read
		if err != nil && err != io.EOF {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	elapsed := time.Since(start)

	if totalRead != dataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", totalRead, dataSize)
	}

	fmt.Printf("Took %v\n", elapsed)
	maxTime := time.Duration(partsAmount-1) * time.Second
	minTime := time.Duration(partsAmount) * time.Second
	if elapsed.Abs().Round(time.Second) < maxTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too quickly, elapsed time: %v < max time: %v", elapsed, maxTime)
	} else if elapsed.Abs().Round(time.Second) > minTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too slow, elapsed time: %v > min time: %v", elapsed, minTime)
	}
}

func TestRateLimitedReader_NoLimitRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	data := strings.Repeat("A", dataSize)
	reader := strings.NewReader(data)

	ratelimitedReader := NewRateLimitedReader(reader, 0)
	buffer := make([]byte, dataSize)

	start := time.Now()
	n, err := ratelimitedReader.Read(buffer)
	elapsed := time.Since(start)

	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != dataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", n, dataSize)
	}

	fmt.Printf("Took %v\n", elapsed)
	if elapsed.Abs().Round(time.Second) != 0 { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too slow, elapsed time: %v, expected time: 0s", elapsed)
	}
}

func TestRateLimitedReader_MultipleReads(t *testing.T) {
	dataSize := 10240 // 10 KB of data
	partsAmount := 2
	partPartsAmount := 3 // times to call read for one part - hitting limit
	data := strings.Repeat("A", dataSize)
	reader := strings.NewReader(data)
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)
	buffer := make([]byte, limit/int64(partPartsAmount))

	var totalRead int
	start := time.Now()

	for totalRead < len(data) {
		read, err := ratelimitedReader.Read(buffer)
		totalRead += read
		if err != nil && err != io.EOF {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	elapsed := time.Since(start)

	if totalRead != dataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", totalRead, dataSize)
	}

	fmt.Printf("Took %v\n", elapsed)
	maxTime := time.Duration(partsAmount-1) * time.Second
	minTime := time.Duration(partsAmount) * time.Second
	if elapsed.Abs().Round(time.Second) < maxTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too quickly, elapsed time: %v < max time: %v", elapsed, maxTime)
	} else if elapsed.Abs().Round(time.Second) > minTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too slow, elapsed time: %v > min time: %v", elapsed, minTime)
	}
}

func TestRateLimitedReader_EOFBehavior(t *testing.T) {
	dataSize := 1024 // 1 KB of data
	data := strings.Repeat("A", dataSize)
	reader := strings.NewReader(data)
	limit := int64(dataSize * 20)

	ratelimitedReader := NewRateLimitedReader(reader, limit)
	buffer := make([]byte, dataSize*2)

	n, err := ratelimitedReader.Read(buffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != dataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", n, dataSize)
	}

	n, err = ratelimitedReader.Read(buffer)
	if err != io.EOF {
		t.Fatalf("expected EOF but got error: %v", err)
	} else if n != 0 {
		t.Fatalf("read bytes after getting EOF: got %d", n)
	}
}
