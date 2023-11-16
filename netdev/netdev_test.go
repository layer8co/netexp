package netdev_test

import (
	"fmt"
	"testing"
	"netexp/netdev"
)

func TestGetTraffic(t *testing.T) {
	data :=
`Inter-|   Receive                                                |  Transmit
  face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: %d         99393    0    0    0     0          0         0    %d      99393    0    0    0     0       0          0
  eth0: %d          2497    0    0    0     0          0         0    %d       5772    0    0    0     0       0          0
enp3s0: %d         92179    0    0    0     0          0         0    %d       8210    0    0    0     0       0          0
 wlan0: %d        169069    0    0    0     0          0         0    %d      98492    0    0    0     0       0          0
docker0:  314460     645    0    0    0     0          0         0   746608     784    0    0    0     0       0          0
veth:     323490     645    0    0    0     0          0         0   767960     854    0    0    0     0       0          0
`

	var recv_lo, trns_lo  int64 = 40979312,   941248120
	var recv_1,  trns_1   int64 = 12818024,   71254138211
	var recv_2,  trns_2   int64 = 5413123928, 95481284
	var recv_3,  trns_3   int64 = 283149218,  112321

	data = fmt.Sprintf(
		data,
		recv_lo, trns_lo,
		recv_1,  trns_1,
		recv_2,  trns_2,
		recv_3,  trns_3,
	)

	recv, trns, err := netdev.GetTraffic([]byte(data))
	if err != nil {
		t.Fatalf("could not parse netdev: %s", err)
	}

	want_recv := recv_1 + recv_2 + recv_3
	if want_recv != recv {
		t.Errorf("incorrect recv; got %d want %d", recv, want_recv)
	}

	want_trns := trns_1 + trns_2 + trns_3
	if want_trns != trns {
		t.Errorf("incorrect trns; got %d want %d", trns, want_trns)
	}
}
