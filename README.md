# percentile_go

A streaming percentile tracker that keeps a bounded, uniformly random sample of a float64 stream and answers p50, p90, and p99 queries.

## Usage

```go
package main

import (
	"fmt"
	"math/rand"

	percentile_go "example.com/percentile_go"
)

func main() {
	tracker := percentile_go.NewTracker(1000, rand.New(rand.NewSource(1)))
	for i := 0; i < 10000; i++ {
		tracker.Add(float64(i))
	}
	fmt.Println(tracker.P50())
	fmt.Println(tracker.P90())
	fmt.Println(tracker.P99())
}
```

## Why this exists

Storing every value in a long-running stream is expensive, but exact percentiles require the full sorted data. This library trades exactness for bounded memory: it keeps at most `capacity` samples chosen by reservoir sampling, so each value seen so far has an equal chance of being retained. Percentile queries then run against that sample. The result is an estimate, not an exact figure, and the error shrinks as the reservoir grows.

## Edge cases

- A `capacity` of zero or less is clamped to one, so a `Tracker` is always usable.
- `Percentile` clamps its argument to `(0, 100]`; `p <= 0` returns the smallest retained sample and `p > 100` returns the largest.
- Calling any percentile method before `Add` returns `0`.
- Passing `nil` as the `*rand.Rand` uses a fixed seed, which makes repeated runs reproducible but not statistically independent.
