// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package netdev

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const data = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 4097124    99393    0    0    0     0          0         0    673328  99393    0    0    0     0       0          0
  eth0: %d          2497    0    0    0     0          0         0    %d       5772    0    0    0     0       0          0
enp3s0: %d         92179    0    0    0     0          0         0    %d       8210    0    0    0     0       0          0
 wlan0: %d        169069    0    0    0     0          0         0    %d      98492    0    0    0     0       0          0
docker0:  314460     645    0    0    0     0          0         0   746608     784    0    0    0     0       0          0
veth:     227428     645    0    0    0     0          0         0   867960     854    0    0    0     0       0          0
`

const (
	recv1, trns1 int64 = 12818024, 71254138211
	recv2, trns2 int64 = 5413123928, 95481284
	recv3, trns3 int64 = 283149218, 112321

	wantRecv = recv1 + recv2 + recv3
	wantTrns = trns1 + trns2 + trns3
)

var r = strings.NewReader(fmt.Sprintf(
	data,
	recv1, trns1,
	recv2, trns2,
	recv3, trns3,
))

var ifaceRegexp = regexp.MustCompile(IfacePattern)

func TestTraffic(t *testing.T) {
	d := New(ifaceRegexp.Match, nil)
	for range 10 {
		r.Seek(0, io.SeekStart)
		recv, trns, err := d.traffic(r)
		assert.NoError(t, err)
		assert.Equal(t, wantRecv, recv)
		assert.Equal(t, wantTrns, trns)
	}
}

func TestTraffic_NoAlloc(t *testing.T) {
	d := New(ifaceRegexp.Match, nil)
	wantAllocs := float64(0)
	allocs := testing.AllocsPerRun(100, func() {
		r.Seek(0, io.SeekStart)
		d.traffic(r)
	})
	assert.Equal(t, wantAllocs, allocs)
}

// [2025-12-25] 5800 ns/op, 0 allocs/op on an 8-core 10th-gen Intel.
func BenchmarkTraffic(b *testing.B) {
	nd := New(ifaceRegexp.Match, nil)
	for b.Loop() {
		r.Seek(0, io.SeekStart)
		nd.traffic(r)
	}
}

func BenchmarkTrafficSystem(b *testing.B) {
	d := New(ifaceRegexp.Match, nil)
	for b.Loop() {
		d.Traffic()
	}
}
