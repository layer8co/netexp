// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package metrics_test

import (
	"bytes"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/layer8co/netexp/internal/metrics"
)

func TestMetrics(t *testing.T) {
	m := metrics.New(metrics.Config{
		Interval: time.Second,
		BurstWindows: []time.Duration{
			1 * time.Second,
			2 * time.Second,
		},
		OutputWindows: []time.Duration{
			5 * time.Second,
			10 * time.Second,
		},
	})
	steps := []struct {
		line      int
		recv      int64
		trns      int64
		wantLines []string
	}{
		{l(), 10, 10, []string{
			"netexp_recv_bytes 10",
			"netexp_trns_bytes 10",
		}},
		{l(), 20, 20, []string{
			"netexp_recv_bytes 20",
			"netexp_trns_bytes 20",
		}},
		{l(), 40, 38, []string{
			"netexp_recv_bytes 40",
			"netexp_trns_bytes 38",
		}},
		{l(), 50, 50, []string{
			"netexp_recv_bytes 50",
			"netexp_trns_bytes 50",
		}},
		{l(), 60, 60, []string{
			"netexp_recv_bytes 60",
			"netexp_trns_bytes 60",
		}},
		{l(), 70, 70, []string{
			"netexp_recv_bytes 70",
			"netexp_trns_bytes 70",
			"netexp_max_1s_recv_burst_bps_over_5s 20",
			"netexp_max_1s_trns_burst_bps_over_5s 18",
		}},
		{l(), 80, 80, []string{
			"netexp_recv_bytes 80",
			"netexp_trns_bytes 80",
			"netexp_max_1s_recv_burst_bps_over_5s 20",
			"netexp_max_1s_trns_burst_bps_over_5s 18",
			"netexp_max_2s_recv_burst_bps_over_5s 15",
			"netexp_max_2s_trns_burst_bps_over_5s 15",
		}},
		{l(), 82, 85, []string{
			"netexp_recv_bytes 82",
			"netexp_trns_bytes 85",
			"netexp_max_1s_recv_burst_bps_over_5s 10",
			"netexp_max_1s_trns_burst_bps_over_5s 12",
			"netexp_max_2s_recv_burst_bps_over_5s 15",
			"netexp_max_2s_trns_burst_bps_over_5s 15",
		}},
		{l(), 90, 90, []string{
			"netexp_recv_bytes 90",
			"netexp_trns_bytes 90",
			"netexp_max_1s_recv_burst_bps_over_5s 10",
			"netexp_max_1s_trns_burst_bps_over_5s 10",
			"netexp_max_2s_recv_burst_bps_over_5s 10",
			"netexp_max_2s_trns_burst_bps_over_5s 11",
		}},
		{l(), 115, 120, []string{
			"netexp_recv_bytes 115",
			"netexp_trns_bytes 120",
			"netexp_max_1s_recv_burst_bps_over_5s 25",
			"netexp_max_1s_trns_burst_bps_over_5s 30",
			"netexp_max_2s_recv_burst_bps_over_5s 16",
			"netexp_max_2s_trns_burst_bps_over_5s 17",
		}},
	}
	for i, s := range steps {
		t.Run(fmt.Sprintf("step%d-line%d", i, s.line), func(t *testing.T) {
			b := m.Step(s.recv, s.trns, nil)
			gotLines := lines(b)
			wantLines := slices.Clone(s.wantLines)
			slices.Sort(gotLines)
			slices.Sort(wantLines)
			diff := lineDiff(wantLines, gotLines)
			if diff != "" {
				t.Errorf("incorrect result (-want +got):\n%s", diff)
			}
		})
	}
}

func lines(b []byte) (s []string) {
	for line := range bytes.SplitSeq(b, []byte{'\n'}) {
		if len(line) > 0 {
			s = append(s, string(line))
		}
	}
	return s
}

func lineDiff(x, y any) string {
	opt := cmpopts.AcyclicTransformer("SplitLines", func(s string) []string {
		return strings.Split(s, "\n")
	})
	return cmp.Diff(x, y, opt)
}

// l returns the line number where it's called.
func l() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}
