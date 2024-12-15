package rateLimitedReader

import (
	"io"
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
	chunkSize := int64(len(p))
	allowedBytes := r.limit

	if chunkSize > allowedBytes {
		p = p[:allowedBytes]
	}

	expectedTime := time.Duration(chunkSize * int64(time.Second) / r.limit)
	elapsed := time.Since(r.lastRead)

	if elapsed < expectedTime {
		time.Sleep(expectedTime - elapsed)
	}

	n, err = r.reader.Read(p)

	if n > 0 {
		r.lastRead = time.Now()
	}

	return n, err
}
