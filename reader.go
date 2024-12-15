package rateLimitedReader

import (
	"io"
	"sync/atomic"
	"time"
)

type RateLimitedReader struct {
	reader   io.Reader
	limit    int64
	lastRead time.Time
}

func NewRateLimitedReader(r io.Reader, limit int64) *RateLimitedReader {
	return &RateLimitedReader{
		reader: r,
		limit:  limit,
	}
}

func (r *RateLimitedReader) Read(p []byte) (n int, err error) {
	var totalRead int64
	chunkSize := int64(len(p))

	for totalRead < chunkSize {
		limit := atomic.LoadInt64(&r.limit)
		allowedBytes := limit

		if chunkSize-totalRead < allowedBytes {
			allowedBytes = chunkSize - int64(totalRead)
		}

		expectedTime := time.Duration(allowedBytes * int64(time.Second) / limit)
		elapsed := time.Since(r.lastRead)

		if elapsed < expectedTime {
			time.Sleep(expectedTime - elapsed)
		}

		n, err = r.reader.Read(p[totalRead:int(totalRead+allowedBytes)])
		if err != nil {
			if err == io.EOF {
				return int(totalRead), nil
			}
			return int(totalRead), err
		}

		totalRead += int64(n)
		r.lastRead = time.Now()
	}

	return int(totalRead), nil
}

func (r *RateLimitedReader) UpdateLimit(newLimit int64) {
	atomic.StoreInt64(&r.limit, newLimit)
}
