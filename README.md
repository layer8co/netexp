# netexp

netexp is a Prometheus exporter that provides advanced network usage metrics.

It provides the amount of `transmitted` and `recieved` bytes in each second,
from the active network interface. besides that, it also provides the maximum `bursts`
of these two qualities, in different time durations.

By default, the information is based on the pseudo-file `/proc/net/dev` which is
populated by the Linux kernel.

## Installation

### Using go toolchain 
```go
go install github.com/layer8co/netexp/cmd/netexp@latest
```

## Usage
```bash
$ netexp --help 

netexp is a Prometheus exporter that provides advanced network usage metrics.

Usage:
  -burst-windows string
    	comma-separated burst window durations (default "1s,5s")
  -iface-regexp string
    	regexp to match network interface names (default "^(eth\\d+|en[osp]\\d+\\S+|enx\\S+|w[lw]\\S+)$")
  -interval duration
    	polling interval (e.g. 500ms, 1s) (default 1s)
  -listen string
    	address to listen on (default ":9298")
  -output-windows string
    	comma-separated output window durations (default "15s,30s,60s")

$ netexp -listen :9290
listening on :9298
matched interfaces: enp0s31f6, wlp4s0
```

## Exported metrics

Here is the example output:
```
netexp_recv_bytes 1443950207
netexp_trns_bytes 192449225

netexp_max_1s_recv_burst_bps_over_15s 11169295
netexp_max_1s_trns_burst_bps_over_15s 148677

netexp_max_1s_recv_burst_bps_over_30s 11169295
netexp_max_1s_trns_burst_bps_over_30s 148677

netexp_max_1s_recv_burst_bps_over_1m0s 11169295
netexp_max_1s_trns_burst_bps_over_1m0s 148677

netexp_max_5s_recv_burst_bps_over_15s 8127323
netexp_max_5s_trns_burst_bps_over_15s 114077

netexp_max_5s_recv_burst_bps_over_30s 8127323
netexp_max_5s_trns_burst_bps_over_30s 114077

netexp_max_5s_recv_burst_bps_over_1m0s 8127323
netexp_max_5s_trns_burst_bps_over_1m0s 114077
```

- `netexp_recv_bytes` The total number of bytes of data, that has been recieved
  by the interface. which is in our case: 1443950207

- `netexp_trns_bytes` The total number of bytes of data, that has been recieved
  by the interface. which is in our case: 192449225
  
- `netexp_max_{burst-duration}_{direction}_burst_bps_over_{observation-duration}`
Shows how much the maximum traffic rate observed within specific time windows.
It basically shows __The Peak Rates__ of the network interface at small time
windows.
