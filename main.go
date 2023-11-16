package main

import (
	"fmt"
	"time"
	"net/http"
	"netexp/pipeline"
	"netexp/netdev"
)

type Metric struct {
	Name string
	Data int64
}

func main() {
	var metrics []byte

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.Write(metrics)
	})

	go http.ListenAndServe(":9098", nil)

	p := pipeline.New([]int{ 1, 5, 10, 15, 30, 60 })

	for {
		data, err := netdev.ReadNetDev()
		if err != nil {
			panic(fmt.Errorf("could not read netdev: %w", err))
		}

		recv, trns, err := netdev.GetTraffic(data)
		if err != nil {
			panic(fmt.Errorf("could not get traffic: %w", err))
		}

		metrics = p.Step(recv, trns)

		time.Sleep(time.Second)
	}
}
