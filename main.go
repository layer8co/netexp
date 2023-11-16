package main

import (
	"fmt"
	"flag"
	"time"
	"net/http"
	"netexp/pipeline"
	"netexp/netdev"
)

var (
	version = "0.3.4"
	metrics []byte
	bind string
)

func main() {
	flag.StringVar(&bind, "bind", ":9298", "network address to listen on")
	printver := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *printver {
		fmt.Println(version)
		return
	}

	serve()
	gather()
}

func serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("netexp " + version))
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request){
		w.Write(metrics)
	})

	go func() {
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			panic(fmt.Errorf("could not serve http: %w", err))
		}
	}()

	fmt.Printf("listening on %s\n", bind)
}

func gather() {
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
