package percentile_go

import (
	"math"
	"math/rand"
	"testing"
)

const eps = 1e-9

func newSeeded(capacity int) *Tracker {
	return NewTracker(capacity, rand.New(rand.NewSource(42)))
}

func TestEmptyTracker(t *testing.T) {
	tr := newSeeded(10)
	if got := tr.Count(); got != 0 {
		t.Fatalf("Count() = %d, want 0", got)
	}
	if got := tr.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if got := tr.P50(); math.Abs(got-0) > eps {
		t.Fatalf("P50() = %v, want 0", got)
	}
	if got := tr.P90(); math.Abs(got-0) > eps {
		t.Fatalf("P90() = %v, want 0", got)
	}
	if got := tr.P99(); math.Abs(got-0) > eps {
		t.Fatalf("P99() = %v, want 0", got)
	}
}

func TestSingleValue(t *testing.T) {
	tr := newSeeded(10)
	tr.Add(7.5)
	if got := tr.P50(); math.Abs(got-7.5) > eps {
		t.Fatalf("P50() = %v, want 7.5", got)
	}
	if got := tr.P90(); math.Abs(got-7.5) > eps {
		t.Fatalf("P90() = %v, want 7.5", got)
	}
	if got := tr.P99(); math.Abs(got-7.5) > eps {
		t.Fatalf("P99() = %v, want 7.5", got)
	}
}

func TestSmallSortedStream(t *testing.T) {
	tr := newSeeded(100)
	for i := 1; i <= 100; i++ {
		tr.Add(float64(i))
	}
	if got := tr.P50(); math.Abs(got-50) > eps {
		t.Fatalf("P50() = %v, want 50", got)
	}
	if got := tr.P90(); math.Abs(got-90) > eps {
		t.Fatalf("P90() = %v, want 90", got)
	}
	if got := tr.P99(); math.Abs(got-99) > eps {
		t.Fatalf("P99() = %v, want 99", got)
	}
}

func TestCapacityOne(t *testing.T) {
	tr := newSeeded(1)
	tr.Add(1)
	tr.Add(2)
	tr.Add(3)
	if got := tr.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
	if got := tr.Count(); got != 3 {
		t.Fatalf("Count() = %d, want 3", got)
	}
	// With capacity 1, the single retained sample is whatever the reservoir
	// last accepted. We only assert it is one of the inputs.
	got := tr.P50()
	if got != 1 && got != 2 && got != 3 {
		t.Fatalf("P50() = %v, want one of 1,2,3", got)
	}
}

func TestCapacityClampedToOne(t *testing.T) {
	tr := NewTracker(0, rand.New(rand.NewSource(1)))
	tr.Add(5)
	if got := tr.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
	if got := tr.P50(); math.Abs(got-5) > eps {
		t.Fatalf("P50() = %v, want 5", got)
	}
}

func TestReservoirBounded(t *testing.T) {
	tr := newSeeded(10)
	for i := 0; i < 1000; i++ {
		tr.Add(float64(i))
	}
	if got := tr.Len(); got != 10 {
		t.Fatalf("Len() = %d, want 10", got)
	}
	if got := tr.Count(); got != 1000 {
		t.Fatalf("Count() = %d, want 1000", got)
	}
}

func TestPercentileClamping(t *testing.T) {
	tr := newSeeded(100)
	for i := 1; i <= 100; i++ {
		tr.Add(float64(i))
	}
	// p <= 0 clamps to the minimum sample.
	if got := tr.Percentile(0); math.Abs(got-1) > eps {
		t.Fatalf("Percentile(0) = %v, want 1", got)
	}
	// p > 100 clamps to the maximum sample.
	if got := tr.Percentile(200); math.Abs(got-100) > eps {
		t.Fatalf("Percentile(200) = %v, want 100", got)
	}
}

func TestPercentileNearestRank(t *testing.T) {
	tr := newSeeded(4)
	for _, v := range []float64{10, 20, 30, 40} {
		tr.Add(v)
	}
	cases := []struct {
		p    float64
		want float64
	}{
		{25, 10},
		{50, 20},
		{75, 30},
		{100, 40},
	}
	for _, c := range cases {
		if got := tr.Percentile(c.p); math.Abs(got-c.want) > eps {
			t.Errorf("Percentile(%v) = %v, want %v", c.p, got, c.want)
		}
	}
}

func TestDuplicateValues(t *testing.T) {
	tr := newSeeded(10)
	for i := 0; i < 5; i++ {
		tr.Add(3.0)
	}
	for i := 0; i < 5; i++ {
		tr.Add(9.0)
	}
	if got := tr.P50(); math.Abs(got-3) > eps {
		t.Fatalf("P50() = %v, want 3", got)
	}
	if got := tr.P90(); math.Abs(got-9) > eps {
		t.Fatalf("P90() = %v, want 9", got)
	}
}

func TestNegativeAndZeroValues(t *testing.T) {
	tr := newSeeded(10)
	for _, v := range []float64{-5, -1, 0, 1, 5} {
		tr.Add(v)
	}
	if got := tr.P50(); math.Abs(got-0) > eps {
		t.Fatalf("P50() = %v, want 0", got)
	}
	if got := tr.P90(); math.Abs(got-5) > eps {
		t.Fatalf("P90() = %v, want 5", got)
	}
}

func TestDeterministicWithSeed(t *testing.T) {
	a := NewTracker(50, rand.New(rand.NewSource(7)))
	b := NewTracker(50, rand.New(rand.NewSource(7)))
	for i := 0; i < 500; i++ {
		v := float64((i * 37) % 101)
		a.Add(v)
		b.Add(v)
	}
	if math.Abs(a.P50()-b.P50()) > eps {
		t.Fatalf("P50 mismatch: %v vs %v", a.P50(), b.P50())
	}
	if math.Abs(a.P90()-b.P90()) > eps {
		t.Fatalf("P90 mismatch: %v vs %v", a.P90(), b.P90())
	}
	if math.Abs(a.P99()-b.P99()) > eps {
		t.Fatalf("P99 mismatch: %v vs %v", a.P99(), b.P99())
	}
}

func TestNilRandIsDeterministic(t *testing.T) {
	a := NewTracker(10, nil)
	b := NewTracker(10, nil)
	for i := 0; i < 100; i++ {
		a.Add(float64(i))
		b.Add(float64(i))
	}
	if math.Abs(a.P50()-b.P50()) > eps {
		t.Fatalf("P50 mismatch: %v vs %v", a.P50(), b.P50())
	}
}
