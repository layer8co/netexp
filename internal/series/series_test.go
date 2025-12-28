// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package series_test

import (
	"testing"
	"time"

	"github.com/layer8co/netexp/internal/series"
	"github.com/stretchr/testify/assert"
)

func TestSeries(t *testing.T) {

	s := series.New(
		1*time.Second,
		5*time.Second,
	)

	s.Put(1)
	s.Put(2)
	s.Put(3)
	s.Put(10)
	s.Put(20)
	s.Put(50)
	s.Put(30)
	s.Put(40)

	wantSamples := []int64{
		3,
		10,
		20,
		50,
		30,
		40,
	}
	assert.Equal(t, wantSamples, s.Samples)

	assert.Panics(t, func() { s.Max(0) })
	assert.Equal(t, int64(40), mustSeries(s.Max(1*time.Second)))
	assert.Equal(t, int64(40), mustSeries(s.Max(2*time.Second)))
	assert.Equal(t, int64(50), mustSeries(s.Max(3*time.Second)))
	assert.Equal(t, int64(50), mustSeries(s.Max(4*time.Second)))
	assert.Equal(t, int64(50), mustSeries(s.Max(5*time.Second)))
	assert.Equal(t, int64(50), mustSeries(s.Max(6*time.Second)))
	_, hasEnoughSamples := s.Max(7 * time.Second)
	assert.Equal(t, false, hasEnoughSamples)

	assert.Panics(t, func() { s.Rate(0) })
	assert.Equal(t, int64(10), mustSeries(s.Rate(1*time.Second)))
	assert.Equal(t, int64(-5), mustSeries(s.Rate(2*time.Second)))
	assert.Equal(t, int64(6), mustSeries(s.Rate(3*time.Second)))
	assert.Equal(t, int64(7), mustSeries(s.Rate(4*time.Second)))
	assert.Equal(t, int64(7), mustSeries(s.Rate(5*time.Second)))
	_, hasEnoughSamples = s.Rate(6 * time.Second)
	assert.Equal(t, false, hasEnoughSamples)
}

func mustSeries[T any](v T, ok bool) T {
	if !ok {
		panic("not enough samples")
	}
	return v
}
