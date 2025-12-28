// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

// TODO:
// - Test netdev.*NetDev.Traffic + $HOST_PROC.
// - Add proper logging.
// - Test netdev's logging.
// - Once layer8co/toolbox/container/ringbuf is ready, use it in netdev for storing samples instead of the []int64.
// - Implement a generic bucketed pool in layer8co/toolbox and use that in rcu.*BufferRcu instead of sync.Pool.
// - Move rcu to layer8co/toolbox.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/layer8co/netexp/internal/metrics"
	"github.com/layer8co/netexp/internal/netdev"
	"github.com/layer8co/netexp/internal/rcu"
)

const (
	appName  = "netexp"
	helpText = "netexp is a Prometheus exporter that provides advanced network usage metrics."
)

var (
	listen = flag.String(
		"listen",
		":9298",
		"address to listen on",
	)
	ifaceRegexpFlag = flag.String(
		"iface-regexp",
		netdev.IfacePattern,
		"regexp to match network interface names",
	)
	interval = flag.Duration(
		"interval",
		time.Second,
		"polling interval (e.g. 500ms, 1s)",
	)
	burstWindowsFlag = flag.String(
		"burst-windows",
		"1s,5s",
		"comma-separated burst window durations",
	)
	outputWindowsFlag = flag.String(
		"output-windows",
		"15s,30s,60s",
		"comma-separated output window durations",
	)
)

var (
	appRcu     = rcu.NewBufferRcu()
	appNetDev  *netdev.NetDev
	appMetrics *metrics.Metrics
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", helpText)
		flag.PrintDefaults()
	}

	flag.Parse()

	ifaceRegexp, err := regexp.Compile(*ifaceRegexpFlag)
	if err != nil {
		die(fmt.Sprintf("-iface-regexp parse erorr: %s", err))
	}

	appNetDev = netdev.New(ifaceRegexp.Match, func(fn func(io.Writer)) {
		b := new(bytes.Buffer)
		fn(b)
		fmt.Printf("%s\n", b.Bytes())
	})

	appMetrics = metrics.New(metrics.Config{
		Interval:      *interval,
		BurstWindows:  mustGet(parseDurations(*burstWindowsFlag)),
		OutputWindows: mustGet(parseDurations(*outputWindowsFlag)),
	})

	fmt.Printf("listening on %s\n", *listen)

	go func() {
		mustDo(gatherMetrics())
	}()

	serveHttp()
}

func serveHttp() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, appName)
	})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		appRcu.Read(func(b []byte) {
			w.Write(b)
		})
	})
	return http.ListenAndServe(*listen, nil)
}

func gatherMetrics() error {
	for ; true; <-time.Tick(*interval) {
		recv, trns, err := appNetDev.Traffic()
		if err != nil {
			return err
		}
		appRcu.Update(func(b []byte) ([]byte, error) {
			b = appMetrics.Step(recv, trns, b)
			b = append(b, '\n')
			return b, nil
		})
	}
	return nil
}

func parseDurations(s string) (out []time.Duration, err error) {
	for field := range strings.SplitSeq(s, ",") {
		field = strings.TrimSpace(field)
		d, err := time.ParseDuration(field)
		if err != nil {
			return nil, fmt.Errorf("could not parse duration %q: %w", field, err)
		}
		out = append(out, d)
	}
	return out, nil
}

func die(s string) {
	fmt.Println(s)
	os.Exit(1)
}

func mustDo(err error) {
	if err != nil {
		die(err.Error())
	}
}

func mustGet[T any](v T, err error) T {
	mustDo(err)
	return v
}
