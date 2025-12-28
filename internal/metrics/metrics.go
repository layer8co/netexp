// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

// Package metrics produces netexp's Prometheus metrics.
package metrics

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"github.com/layer8co/netexp/internal/netdev"
	"github.com/layer8co/netexp/internal/series"
)

// e.g. a burst window of 5 seconds and an output window of 60 seconds
// provides the metric of `the maximum 5 second burst over the last 60 seconds`.
type Metrics struct {
	Config
	recv      *series.TimeSeries
	trns      *series.TimeSeries
	recvBurst []*series.TimeSeries
	trnsBurst []*series.TimeSeries
	netdev    *netdev.NetDev
}

type Config struct {
	Interval      time.Duration
	BurstWindows  []time.Duration
	OutputWindows []time.Duration
}

func New(c Config) *Metrics {
	m := &Metrics{
		Config: c,
	}
	window := slices.Max(m.OutputWindows)
	m.recv = series.New(m.Interval, window)
	m.trns = series.New(m.Interval, window)
	for range m.BurstWindows {
		m.recvBurst = append(m.recvBurst, series.New(m.Interval, window))
		m.trnsBurst = append(m.trnsBurst, series.New(m.Interval, window))
	}
	return m
}
func (m *Metrics) Step(recv, trns int64, b []byte) []byte {
	m.recv.Put(recv)
	m.trns.Put(trns)
	b = fmt.Appendf(b, "netexp_recv_bytes %d\n", recv)
	b = fmt.Appendf(b, "netexp_trns_bytes %d\n", trns)
	for i, bw := range m.BurstWindows {
		recvBurst, ok := m.recv.Rate(bw)
		if ok {
			m.recvBurst[i].Put(recvBurst)
		}
		trnsBurst, ok := m.trns.Rate(bw)
		if ok {
			m.trnsBurst[i].Put(trnsBurst)
		}
		for _, ow := range m.OutputWindows {
			maxRecvBurst, ok := m.recvBurst[i].Max(ow)
			if ok {
				b = fmt.Appendf(
					b,
					"netexp_max_%s_recv_burst_bps_over_%s %d\n",
					bw, ow, maxRecvBurst,
				)
			}
			maxTransBurst, ok := m.trnsBurst[i].Max(ow)
			if ok {
				b = fmt.Appendf(
					b,
					"netexp_max_%s_trns_burst_bps_over_%s %d\n",
					bw, ow, maxTransBurst,
				)
			}
		}
	}
	b = bytes.TrimRight(b, "\n")
	return b
}
