package rateLimitedReader

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestRateLimitedReader_BasicRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 4
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)

	start := time.Now()
	read(t, ratelimitedReader, dataSize, dataSize)
	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
}

func TestRateLimitedReader_NoLimitRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	reader := bytes.NewReader(make([]byte, dataSize))

	ratelimitedReader := NewRateLimitedReader(reader, 0)

	start := time.Now()
	read(t, ratelimitedReader, dataSize, dataSize)
	assertReadTimes(t, time.Since(start), 0, 0)
}

func TestRateLimitedReader_MultipleReads(t *testing.T) {
	dataSize := 10240 // 10 KB of data
	partsAmount := 2
	partPartsAmount := 4 // times to call read for one part - hitting limit
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := dataSize / partsAmount // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, int64(limit))

	var totalRead int
	start := time.Now()

	for totalRead < dataSize {
		// note that there's a validation of the read size - meaning dataSize / partsAmount % partPartsAmount must be 0 (divided without leftovers)
		totalRead += read(t, ratelimitedReader, limit/partPartsAmount, limit/partPartsAmount)
	}

	if totalRead != dataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", totalRead, dataSize)
	}

	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
}

func TestRateLimitedReader_EOFBehavior(t *testing.T) {
	dataSize := 1024 // 1 KB of data
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize * 20)

	ratelimitedReader := NewRateLimitedReader(reader, limit)

	read(t, ratelimitedReader, dataSize*2, dataSize)
	read(t, ratelimitedReader, dataSize*2, 0)
}

func TestRateLimitedReader_UpdateLimit(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 4
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)

	start := time.Now()
	read(t, ratelimitedReader, dataSize/2, dataSize/2)

	ratelimitedReader.UpdateLimit(limit * 2) // update limit to cut time for the second half by half (minus 25% to the expected time)

	read(t, ratelimitedReader, dataSize/2, dataSize/2)

	assertReadTimes(t, time.Since(start), int(float64(partsAmount)*0.75), int(float64(partsAmount)*0.75)+1)
}

func TestRateLimitedReader_GetCurrentTotalRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 4
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)
	doneC := make(chan struct{}, 0)

	go func() {
		defer func() { doneC <- struct{}{} }()
		for i := 0; i < partsAmount; i++ {
			select {
			case <-time.After(time.Second):
				currentTotalRead := ratelimitedReader.GetCurrentTotalRead()
				fmt.Printf("Total Read: %d , LimitAbs: %d\n", currentTotalRead, limit)
				if currentTotalRead != limit*int64(i+1) {
					t.Fatalf("got unexpected CurrentTotalRead, read: %d expected: %d", currentTotalRead, limit*int64(i+1))
				}
			}
		}
	}()

	start := time.Now()
	read(t, ratelimitedReader, dataSize, dataSize)
	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
	<-doneC
}

func TestRateLimitedReader_UnconventionalLimitRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 2
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize/partsAmount - 3000) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)

	start := time.Now()
	read(t, ratelimitedReader, dataSize, dataSize)
	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
}

func TestRateLimitedReader_CopyRead(t *testing.T) {
	dataSize := 102400 // 100 KB of data
	partsAmount := 2
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := int64(dataSize / partsAmount) // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, limit)

	start := time.Now()
	n, err := io.Copy(io.Discard, ratelimitedReader)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != int64(dataSize) {
		t.Fatalf("read incomplete data, read: %d expected: %d", n, dataSize)
	}

	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
}

type mockReadCloser struct {
	closed bool
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

func TestRateLimitedReadCloser_Close(t *testing.T) {
	readCloser := mockReadCloser{}
	ratelimitedReadCloser := NewRateLimitedReadCloser(&readCloser, 0)
	err := ratelimitedReadCloser.Close()
	if err != nil {
		t.Fatalf("unexpected error while closing: %v", err)
	}

	if readCloser.closed == false {
		t.Fatalf("expected readCloser to be closed but wasn't")
	}
}

func TestRateLimitedReader_LargeRead(t *testing.T) {
	dataSize := 1 * 1024 * 1024 * 1024 // 1 GB of data
	partsAmount := 4
	reader := bytes.NewReader(make([]byte, dataSize))
	limit := dataSize / partsAmount // dataSize/partsAmount bytes per second

	ratelimitedReader := NewRateLimitedReader(reader, int64(limit))

	start := time.Now()
	read(t, ratelimitedReader, dataSize, dataSize)
	assertReadTimes(t, time.Since(start), partsAmount, partsAmount+1)
}

func TestRateLimitedReader_ReadPerformence(t *testing.T) {
	const durationInSeconds = 10
	const bufferSize = 32 * 1024 // 32KB buffer
	const limit = bufferSize * 1000
	fmt.Printf("Duration set: %d seconds\n", durationInSeconds)

	buffer := make([]byte, bufferSize)
	var totalBytes int64

	reader := infiniteReader{}
	ratelimitedReader := NewRateLimitedReader(reader, limit) // large limit - no limit
	deadline := time.Now().Add(durationInSeconds * time.Second)

	for time.Now().Before(deadline) {
		n, err := ratelimitedReader.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			break
		}
	}

	deviation := 0.8
	if totalBytes < int64(limit*durationInSeconds*deviation) {
		t.Fatalf("read incomplete data, read: %d expected: %d, with %.2f deviation: %d", totalBytes, limit*durationInSeconds, deviation, int64(limit*durationInSeconds*deviation))
	}

	mb := float64(totalBytes) / 1024.0 / 1024.0
	fmt.Printf("MaxReadOverTimeSyntheticTest: Read %.4f MB in 10 seconds\n", mb)
}

type infiniteReader struct{}

func (infiniteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'A'
	}
	return len(p), nil
}

func read(t *testing.T, reader *RateLimitedReader, bufferSize, expectedDataSize int) int {
	buffer := make([]byte, bufferSize)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != expectedDataSize {
		t.Fatalf("read incomplete data, read: %d expected: %d", n, expectedDataSize)
	}

	return n
}

func assertReadTimes(t *testing.T, elapsed time.Duration, minTimeInSeconds, maxTimeInSeconds int) {
	fmt.Printf("Took %v\n", elapsed)
	minTime := time.Duration(minTimeInSeconds) * time.Second
	maxTime := time.Duration(maxTimeInSeconds) * time.Second
	if elapsed.Abs().Round(time.Second) < minTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too quickly, elapsed time: %v < min time: %v", elapsed, minTime)
	} else if elapsed.Abs().Round(time.Second) > maxTime { // round to second - has a deviation of up to half a second
		t.Errorf("read completed too slow, elapsed time: %v > max time: %v", elapsed, maxTime)
	}
}
