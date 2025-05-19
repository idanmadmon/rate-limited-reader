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
    "os"
    "github.com/idanmadmon/rate-limited-reader"
)

func main() {
    file, _ := os.Open("big_video.mp4")

    // Limit to 1 Megabyte/sec
    r := ratelimitedreader.New(file, ratelimitedreader.Config{
        RateBytesPerSec: 1024 * 1024,
    })

    // use `r` as any io.Reader
    buf := make([]byte, 4096)
    for {
        n, err := r.Read(buf)
        if n > 0 {
            // do something with buf[:n]
        }
        if err != nil {
            break
        }
    }
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
