package main

import (
	"flag"
	"fmt"
	"net/http"
	"netexp/netdev"
	"netexp/pipeline"
	"netexp/rcu"
	"os"
	"time"
)

var (
	version = "0.3.8"
	metrics rcu.RCU[[]byte]
	listen  string
	getver  bool
)

func main() {
	// get options from flags
	flag.StringVar(&listen, "listen", ":9298", "network address to listen on")
	flag.BoolVar(&getver, "version", false, "print version and exit")

	// get options from env vars
	env := os.Getenv("NETEXP_LISTEN")
	if env != "" {
		listen = env
	}

	flag.Parse()

	if getver {
		fmt.Println(version)
		return
	}

	serve()
	gather()
}

func serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("netexp " + version + "\n"))
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Write(*metrics.Load())
	})

	go func() {
		err := http.ListenAndServe(listen, nil)
		if err != nil {
			panic(fmt.Errorf("could not serve http: %w", err))
		}
	}()

	fmt.Printf("listening on %s\n", listen)
}

func gather() {
	p := pipeline.New([]int{1, 5, 10, 15, 30, 60})

	for {
		data, err := netdev.ReadNetDev()
		if err != nil {
			panic(fmt.Errorf("could not read netdev: %w", err))
		}

		recv, trns, err := netdev.GetTraffic(data)
		if err != nil {
			panic(fmt.Errorf("could not get traffic: %w", err))
		}

		buf := p.Step(recv, trns)

		metrics.Store(&buf)

		time.Sleep(time.Second)
	}
}
