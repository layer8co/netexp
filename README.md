# ⇄ netexp

netexp is a Prometheus exporter that provides advanced network usage metrics.

It provides the amount of `transmitted` and `recieved` bytes in each second,
from the active network interface. besides that, it also provides the maximum `bursts`
of these two qualities, in different time durations.

By default, the information is based on the pseudo-file `/proc/net/dev` which is
populated by the Linux kernel.



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

- `netexp_recv_bytes` The total number of bytes of data, that has been recieved
  by the interface

- `netexp_trns_bytes` The total number of bytes of data, that has been recieved
  by the interface
  
- `netexp_max_{burst-duration}_{direction}_burst_bps_over_{observation-duration}`
Shows how much the maximum traffic rate observed within specific time windows.
It basically shows the __The Peak Rates__ of the network interface at small time
windows.


