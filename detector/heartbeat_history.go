package detector

import (
	"errors"
	"math"
	"sync"
)

// HeartbeatHistory represents a sample window of finite capacity
type HeartbeatHistory struct {
	// MaxSamples is the maximum number of samples for the window
	MaxSamples int
	// Intervals is a slice of intervals
	Intervals          []float64
	intervalSum        float64
	squaredIntervalSum float64
	sync.Mutex
}

// NewHeartbeatHistory creates a new HeartbeatHistory
func NewHeartbeatHistory(maxSamples int) (*HeartbeatHistory, error) {
	if maxSamples <= 0 {
		return nil, errors.New("sample capacity must be > 0")
	}

	hh := &HeartbeatHistory{
		MaxSamples: maxSamples,
	}
	return hh, nil
}

// Mean returns the mean average of the intervals
func (hh *HeartbeatHistory) Mean() float64 {
	return hh.intervalSum / float64(len(hh.Intervals))
}

// Variance returns the variances of the interval distribution
func (hh *HeartbeatHistory) Variance() float64 {
	return (hh.squaredIntervalSum / float64(len(hh.Intervals))) - (hh.Mean() * hh.Mean())
}

// StdDev returns the standard deviation of the interval distribution
func (hh *HeartbeatHistory) StdDev() float64 {
	return math.Sqrt(hh.Variance())
}

// DropOldestInterval drops the oldest interval in the history
func (hh *HeartbeatHistory) DropOldestInterval() {
	hh.Lock()
	intervalToDrop := float64(hh.Intervals[0])
	hh.Intervals = hh.Intervals[1:]
	hh.intervalSum = hh.intervalSum - intervalToDrop
	hh.squaredIntervalSum = hh.squaredIntervalSum - (intervalToDrop * intervalToDrop)
	hh.Unlock()
}

// AddInterval adds a new interval to the slice of intervals
func (hh *HeartbeatHistory) AddInterval(interval float64) {
	if len(hh.Intervals) < hh.MaxSamples {
		hh.Lock()
		hh.Intervals = append(hh.Intervals, interval)
		hh.intervalSum = hh.intervalSum + float64(interval)
		hh.squaredIntervalSum = hh.squaredIntervalSum + (float64(interval) * float64(interval))
		hh.Unlock()
	} else {
		hh.DropOldestInterval()
		hh.AddInterval(interval)
	}
}
