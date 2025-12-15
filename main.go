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
	listen  string
	getver  bool
)

type Metrics []byte

var (
	rcuMetrics *rcu.RCU[Metrics]
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

	// Init rcu
	rcuMetrics = rcu.New[Metrics]()
	// Add the first element.
	rcuMetrics.Rotate()

	serve()
	gather()
}

func serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("netexp " + version + "\n"))
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		latest, done := rcuMetrics.Latest()

		if latest != nil {
			w.Write(*latest)
			done()
		}
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

	timer := time.NewTicker(60 * time.Second)

	for {
		data, err := netdev.ReadNetDev()
		if err != nil {
			panic(fmt.Errorf("could not read netdev: %w", err))
		}

		recv, trns, err := netdev.GetTraffic(data)
		if err != nil {
			panic(fmt.Errorf("could not get traffic: %w", err))
		}

		m := p.Step(recv, trns)

		// Non blocking. It expected to be fast.
		rcuMetrics.Assign(m)

		select {
		case <-timer.C:
			rcuMetrics.Rotate()
		default:
		}

		time.Sleep(time.Second)
	}
}
