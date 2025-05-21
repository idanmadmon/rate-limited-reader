# rate-limited-reader

**Deterministic, real-time friendly rate limiting for Go readers**

Built for systems that care about _tempo_, _predictability_, and _adaptability_.

</br>


## Why?

Traditional rate limiters (like token buckets) are great for handling bursts — but in real-time, bandwidth-sensitive systems, you sometimes need something stricter, more _predictable_.
</br>
This library wraps your `io.Reader` and ensures a consistent read rate over time, with optional support for **adaptive control** based on bandwidth constraints.

If you care about:
- **Real-time stability**
- **Network fairness**
- **Controlled data streaming**

This is for you.

</br>


## Features

- **Deterministic rate limiting** (not burstable)

- Simple io.Reader wrapping

- Supports dynamic rate adjustments

- Designed for smart bandwidth control systems

- Lightweight & dependency-free (just Go stdlib)

</br>


## Installation

```bash
go get github.com/idanmadmon/rate-limited-reader
```

</br>

## Usage

```Go
package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	ratelimitedreader "github.com/idanmadmon/rate-limited-reader"
)

func main() {
	const dataSize = 32 * 1024
	reader := bytes.NewBuffer(make([]byte, dataSize))

	// allow 1/4 of the data size per second,
	// should take 32 / (32/4) = 4 seconds
	// reads interval divided evenly by
	// ratelimitedreader.ReadIntervalMilliseconds
	limitedReader := ratelimitedreader.NewRateLimitedReader(reader, dataSize/4)

	var total int
	buffer := make([]byte, 1024)
	start := time.Now()
	for {
		n, err := limitedReader.Read(buffer)
		total += n
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error: %v\n", err)
			}
			break
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Total: %d, Elapsed: %s\n", total, elapsed)
}
```

</br>


## When should you use it?

You want **smooth data streaming** with no traffic spikes

You're feeding data into a **real-time processor**, encoder, or network stream

You're building systems that adapt their throughput over time (e.g., **smart bandwidth control**)

You're tired of fiddling with `golang.org/x/time/rate` just to get stable pacing

</br>


## Benchmarks

Want to see how it stacks up against other Go libraries like `golang.org/x/time/rate`, `uber-go/ratelimit`, or `juju/ratelimit`?
</br>
Check out the `comparison & benchmark article here` — complete with graphs, scenarios, and flame.

</br>


## Looking for automatic bandwidth adjustment?

Check out `smart-bandwidth-control` — a library built on top of `rate-limited-reader`, designed for real-time adaptive systems.
</br>
It **dynamically probes bandwidth** and adjusts read rates accordingly, with a focus on reliability and minimal latency overhead.

This is especially useful for:
- IoT edge devices
- Real-time data ingestion pipelines
- Video/audio streamers
- Any system that must adapt to **changing network conditions**


</br>

## License

MIT. Use it, fork it, build cool stuff.

</br>


## Author

Maintained by @idanmadmon
</br>
Got ideas? Want to integrate this into a bandwidth control stack? Open an issue or reach out.
