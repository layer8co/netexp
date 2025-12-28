// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package series

import (
	"slices"
	"time"
)

type TimeSeries struct {
	Samples  []int64
	Interval time.Duration

	maxSamples int
}

func New(interval, window time.Duration) *TimeSeries {
	if interval <= 0 {
		panic("series.New: interval <= 0")
	}
	if window <= 0 {
		panic("series.New: window <= 0")
	}
	if window%interval != 0 {
		panic("series.New: window is not a multiple of interval")
	}
	maxIntervals := int(window / interval)
	maxSamples := maxIntervals + 1
	return &TimeSeries{
		Samples:    make([]int64, 0, maxSamples),
		maxSamples: maxSamples,
		Interval:   interval,
	}
}

func (s *TimeSeries) Put(sample int64) {
	if len(s.Samples) < s.maxSamples {
		s.Samples = append(s.Samples, sample)
		return
	}
	copy(s.Samples, s.Samples[1:])
	s.Samples[len(s.Samples)-1] = sample
}

// Rate is only applicaple to cumulative series.
// It returns the rate of change per second over the last d duration.
//
// Notes:
//   - d must be >= s.Interval.
//   - d == 0 always returns diff == 0.
//   - d is floored to the nearest multiple of s.interval.
func (s *TimeSeries) Rate(d time.Duration) (diff int64, hasEnoughSamples bool) {
	if d < s.Interval {
		panic("TimeSeries.Diff: duration must be at least one interval")
	}
	intervals := int(d / s.Interval)
	if intervals >= len(s.Samples) {
		return 0, false
	}
	new := s.Samples[len(s.Samples)-1]
	old := s.Samples[len(s.Samples)-1-intervals]
	return int64(float64(new-old) / d.Seconds()), true
}

// Max returns the largest sample in the last d duration.
//
// Notes:
//   - d must be >= s.Interval.
//   - d is floored to the nearest multiple of s.interval.
func (s *TimeSeries) Max(d time.Duration) (max int64, hasEnoughSamples bool) {
	if d < s.Interval {
		panic("TimeSeries.Max: duration must be at least one interval")
	}
	samples := int(d / s.Interval)
	if samples > len(s.Samples) {
		return 0, false
	}
	return slices.Max(s.Samples[len(s.Samples)-samples:]), true
}
