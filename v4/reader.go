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
	reader    io.ReadCloser
	limit     int64
	lastRead  time.Time
	totalRead int64
}

func NewRateLimitedReader(r io.Reader, limit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader: io.NopCloser(r),
		limit:  limit,
	}
}

func NewRateLimitedReadCloser(r io.ReadCloser, limit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader: r,
		limit:  limit,
	}
}

func (r *RateLimitedReader) Read(p []byte) (n int, err error) {
	var totalRead int64
	atomic.StoreInt64(&r.totalRead, totalRead)
	chunkSize := int64(len(p))

	for totalRead < chunkSize {
		totalRead = atomic.LoadInt64(&r.totalRead)
		limit := atomic.LoadInt64(&r.limit)

		// the limit set to per second
		limit = limit / (1000 / ReadIntervalMilliseconds)

		if limit == 0 {
			limit = chunkSize
		}

		allowedBytes := limit

		if chunkSize-totalRead < allowedBytes {
			allowedBytes = chunkSize - totalRead
		}

		expectedTime := time.Duration(allowedBytes * ReadIntervalMilliseconds * int64(time.Millisecond) / limit)
		elapsed := time.Since(r.lastRead)

		if elapsed < expectedTime {
			time.Sleep(expectedTime - elapsed)
		}

		r.lastRead = time.Now()
		n, err = r.reader.Read(p[totalRead:int(totalRead+allowedBytes)])
		if err != nil {
			if err == io.EOF {
				if totalRead == 0 {
					return 0, err
				}
				return int(totalRead), nil
			}
			return int(totalRead), err
		}

		atomic.AddInt64(&r.totalRead, int64(n))
	}

	return int(atomic.LoadInt64(&r.totalRead)), nil
}

func (r *RateLimitedReader) Close() error {
	return r.reader.Close()
}

func (r *RateLimitedReader) UpdateLimit(newLimit int64) {
	atomic.StoreInt64(&r.limit, newLimit)
}

func (r *RateLimitedReader) GetCurrentTotalRead() int64 {
	return atomic.LoadInt64(&r.totalRead)
}
