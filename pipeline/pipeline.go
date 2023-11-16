package pipeline

import (
	"fmt"
	"netexp/series"
	"netexp/math"
)

type Pipeline struct {
	ranges []int
	trns_series *series.Series[int64]
	recv_series *series.Series[int64]
	trns_rates_series map[int]*series.Series[int64]
	recv_rates_series map[int]*series.Series[int64]
}

func New(ranges []int) *Pipeline {
	p := &Pipeline{}

	p.ranges = ranges

	var max int
	for i, r := range ranges {
		if i == 0 || r > max {
			max = r
		}
	}

	p.trns_series = series.New[int64](max + 1)
	p.recv_series = series.New[int64](max + 1)

	p.trns_rates_series = make(map[int]*series.Series[int64])
	p.recv_rates_series = make(map[int]*series.Series[int64])

	for _, r := range ranges {
		p.trns_rates_series[r] = series.New[int64](max)
		p.recv_rates_series[r] = series.New[int64](max)
	}

	return p
}

func (p *Pipeline) Step(recv, trns int64) []byte {
	metrics := make([]byte, 0)

	register := func(name string, data int64) {
		metrics = append(metrics, []byte(fmt.Sprintf("%s %d\n", name, data))...)
	}

	p.recv_series.Record(recv)
	p.trns_series.Record(trns)

	for _, r := range p.ranges {
		if p.trns_series.Length() >= r + 1 {
			trns_rate := math.Rate(p.trns_series.Samples, r + 1)
			recv_rate := math.Rate(p.recv_series.Samples, r + 1)

			p.trns_rates_series[r].Record(trns_rate)
			p.recv_rates_series[r].Record(recv_rate)

			trns_name := fmt.Sprintf("netexp_transmit_rate_%ds_bps", r)
			recv_name := fmt.Sprintf("netexp_receive_rate_%ds_bps",  r)
			register(trns_name, trns_rate)
			register(recv_name, recv_rate)
		}

		for _, m := range p.ranges {
			if m > r && p.trns_rates_series[r].Length() >= m {
				trns_name := fmt.Sprintf("netexp_transmit_rate_%ds_max_%ds_bps", r, m)
				recv_name := fmt.Sprintf("netexp_receive_rate_%ds_max_%ds_bps",  r, m)
				register(trns_name, math.Max(p.trns_rates_series[r].Samples, m))
				register(recv_name, math.Max(p.recv_rates_series[r].Samples, m))
			}
		}
	}

	return metrics
}
