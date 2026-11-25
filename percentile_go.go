// Package percentile_go implements a streaming percentile tracker backed by a
// sorted sample reservoir. It answers p50, p90, and p99 queries over a stream
// of float64 values without storing the entire stream.
//
// The reservoir keeps a fixed-capacity sorted slice of samples. When the
// reservoir is full, a new value replaces a random existing sample with
// probability capacity/seen. This is the classic reservoir sampling algorithm,
// which gives each value seen so far an equal chance of being in the reservoir.
// Percentiles are read from the sorted reservoir using nearest-rank.
package percentile_go

import (
	"math/rand"
	"sort"
)

// Tracker maintains a bounded, uniformly random sample of a stream of float64
// values and answers percentile queries against that sample.
type Tracker struct {
	capacity int
	samples  []float64
	seen     int64
	rng      *rand.Rand
}

// NewTracker returns a Tracker that will retain at most capacity samples.
//
// A capacity of zero or less is treated as one so that a Tracker is always
// usable. The rng argument is used for reservoir replacement decisions; pass
// nil to use a deterministic seed. Tests should pass a seeded *rand.Rand so
// that behaviour does not depend on global random state.
func NewTracker(capacity int, rng *rand.Rand) *Tracker {
	if capacity < 1 {
		capacity = 1
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}
	return &Tracker{
		capacity: capacity,
		samples:  make([]float64, 0, capacity),
		rng:      rng,
	}
}

// Add records a value from the stream. The value may or may not be retained
// in the reservoir, depending on the sampling algorithm.
func (t *Tracker) Add(v float64) {
	t.seen++
	if len(t.samples) < t.capacity {
		t.insertSorted(v)
		return
	}
	// Reservoir sampling: replace a uniformly chosen existing sample with
	// probability capacity/seen. Equivalent to choosing an index in [0, seen)
	// and replacing that index if it falls inside the reservoir.
	idx := t.rng.Int63n(t.seen)
	if idx < int64(t.capacity) {
		t.samples[idx] = v
		sort.Float64s(t.samples)
	}
}

// insertSorted inserts v into the sorted samples slice in O(n) time. The
// reservoir is small, so a simple insertion is faster than appending and
// re-sorting for every early value.
func (t *Tracker) insertSorted(v float64) {
	i := sort.SearchFloat64s(t.samples, v)
	t.samples = append(t.samples, 0)
	copy(t.samples[i+1:], t.samples[i:])
	t.samples[i] = v
}

// Count returns the number of values added to the tracker, including those
// that were not retained in the reservoir.
func (t *Tracker) Count() int64 {
	return t.seen
}

// Len returns the current number of retained samples. It is never greater
// than the tracker's capacity.
func (t *Tracker) Len() int {
	return len(t.samples)
}

// Percentile returns the nearest-rank percentile of the retained samples.
//
// p must be in (0, 100]. Values outside that range are clamped to the nearest
// valid boundary. If no samples have been added, Percentile returns 0.
//
// Nearest-rank is used because it is simple, exact for the retained sample,
// and avoids the interpolation ambiguity that plagues percentile definitions.
func (t *Tracker) Percentile(p float64) float64 {
	if len(t.samples) == 0 {
		return 0
	}
	if p <= 0 {
		p = 0.0001
	}
	if p > 100 {
		p = 100
	}
	n := len(t.samples)
	// Nearest-rank: rank = ceil(p/100 * n), 1-based.
	rank := int(mathCeil(p/100*float64(n)))
	if rank < 1 {
		rank = 1
	}
	if rank > n {
		rank = n
	}
	return t.samples[rank-1]
}

// P50 returns the median of the retained samples.
func (t *Tracker) P50() float64 { return t.Percentile(50) }

// P90 returns the 90th percentile of the retained samples.
func (t *Tracker) P90() float64 { return t.Percentile(90) }

// P99 returns the 99th percentile of the retained samples.
func (t *Tracker) P99() float64 { return t.Percentile(99) }

// mathCeil is a tiny wrapper so the package does not need to import math just
// for one call. It returns the smallest integer value greater than or equal
// to x.
func mathCeil(x float64) float64 {
	i := float64(int64(x))
	if i < x {
		return i + 1
	}
	return i
}
