package rateLimitedReader

import (
	"io"
	"sync/atomic"
	"time"
)

var (
	ReadIntervalMilliseconds int64 = 50
)

type RateLimitedReader struct {
	limit     atomic.Int64
	totalRead atomic.Int64
	reader    io.ReadCloser
	lastRead  time.Time
}

func NewRateLimitedReader(r io.Reader, limit int64) *RateLimitedReader {
	return NewRateLimitedReadCloser(io.NopCloser(r), limit)
}

func NewRateLimitedReadCloser(r io.ReadCloser, limit int64) *RateLimitedReader {
	reader := &RateLimitedReader{
		reader: r,
	}

	reader.limit.Store(limit)
	return reader
}

func (r *RateLimitedReader) Read(p []byte) (n int, err error) {
	r.totalRead.Store(0)
	chunkSize := int64(len(p))
	for r.totalRead.Load() < chunkSize {
		limit := r.limit.Load()
		if limit <= 0 {
			n, err = r.readWithoutLimit(p[r.totalRead.Load():int(chunkSize)])
			r.totalRead.Add(int64(n))
			return int(r.totalRead.Load()), err
		}

		// the limit set to per second
		limit = limit / (1000 / ReadIntervalMilliseconds)

		allowedBytes := limit
		chunkSizeLeft := chunkSize - r.totalRead.Load()
		if chunkSizeLeft < allowedBytes {
			allowedBytes = chunkSizeLeft
		}

		expectedTime := time.Duration(allowedBytes * ReadIntervalMilliseconds * int64(time.Millisecond) / limit)
		elapsed := time.Since(r.lastRead)

		if elapsed < expectedTime {
			time.Sleep(expectedTime - elapsed)
		}

		r.lastRead = time.Now()
		n, err = r.reader.Read(p[r.totalRead.Load():int(r.totalRead.Load()+allowedBytes)])
		r.totalRead.Add(int64(n))
		if err != nil {
			break
		}
	}

	return int(r.totalRead.Load()), err
}

func (r *RateLimitedReader) readWithoutLimit(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *RateLimitedReader) Close() error {
	return r.reader.Close()
}

func (r *RateLimitedReader) UpdateLimit(newLimit int64) {
	r.limit.Store(newLimit)
}

func (r *RateLimitedReader) GetCurrentTotalRead() int64 {
	return r.totalRead.Load()
}
