// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

// Package netdev provides functionality for parsing /proc/net/dev.
package netdev

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/layer8co/toolbox/oslite"
)

const (
	// Pattern for matching typical wide area network interface names.
	//   - https://www.thomas-krenn.com/en/wiki/Predictable_Network_Interface_Names
	//   - https://www.freedesktop.org/software/systemd/man/255/systemd.net-naming-scheme.html
	IfacePattern = `^(eth\d+|en[osp]\d+\S+|enx\S+|w[lw]\S+)$`

	// 1-indexed
	netdevFirstLine = 3

	// 0-indexed
	netdevIfaceField = 0
	netdevRecvField  = 1
	netdevTrnsField  = 9
	netdevMaxField   = 9

	netdevMaxLineSize   = 1024
	ifaceListInitialCap = 128
)

var (
	netdevName = "${HOST_PROC:-/proc}/net/dev"
	netdevPath string

	ifaceListDelim = []byte(", ")
)

func init() {
	hostProc := os.Getenv("HOST_PROC")
	if hostProc == "" {
		hostProc = "/proc"
	}
	netdevPath = hostProc + "/net/dev"
}

type NetDev struct {
	ifaceMatcher MatchFunc
	logger       LogFunc

	ifaceList     []byte
	prevIfaceList []byte

	scanBuf []byte
	file    *oslite.File
}

type (
	MatchFunc func(ifaceName []byte) bool
	LogFunc   func(func(io.Writer))
)

func New(ifaceMatcher MatchFunc, logger LogFunc) *NetDev {
	d := &NetDev{
		ifaceMatcher: ifaceMatcher,
		logger:       logger,
		scanBuf:      make([]byte, netdevMaxLineSize),
		file:         new(oslite.File),
	}
	if d.logger != nil {
		d.ifaceList = make([]byte, 0, ifaceListInitialCap)
		d.prevIfaceList = make([]byte, 0, ifaceListInitialCap)
	}
	return d
}

func (d *NetDev) Traffic() (recv, trns int64, err error) {
	err = d.file.Open(netdevPath)
	if err != nil {
		return 0, 0, fmt.Errorf("could not open file %q: %w", netdevName, err)
	}
	defer d.file.Close()
	return d.traffic(d.file)
}

func (d *NetDev) traffic(r io.Reader) (recv, trns int64, err error) {

	scanner := bufio.NewScanner(r)
	scanner.Buffer(d.scanBuf, cap(d.scanBuf))

	lineNum := 0

	for scanner.Scan() {

		lineNum++
		line := scanner.Bytes()

		if lineNum < netdevFirstLine || len(line) == 0 {
			continue
		}

		fields := make([][]byte, netdevMaxField+1)
		n := readFields(line, fields)
		fields = fields[:n]

		iface := fields[netdevIfaceField]
		iface = bytes.TrimRight(iface, ":")
		recvText := fields[netdevRecvField]
		trnsText := fields[netdevTrnsField]

		if !d.ifaceMatcher(iface) {
			continue
		}

		if d.logger != nil {
			d.ifaceList = append(d.ifaceList, iface...)
			d.ifaceList = append(d.ifaceList, ifaceListDelim...)
		}

		x, err := strconv.ParseInt(string(recvText), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse recv number: %w", err)
		}
		recv += x

		x, err = strconv.ParseInt(string(trnsText), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("could not parse trnx number: %w", err)
		}
		trns += x
	}

	if d.logger != nil {
		if !bytes.Equal(d.ifaceList, d.prevIfaceList) {
			d.logger(func(w io.Writer) {
				fmt.Fprintf(w,
					"matched interfaces: %s",
					bytes.TrimSuffix(d.ifaceList, ifaceListDelim),
				)
			})
		}
		tmp := d.prevIfaceList
		d.prevIfaceList = d.ifaceList
		d.ifaceList = tmp[:0]
	}

	err = scanner.Err()
	if err != nil {
		return 0, 0, fmt.Errorf("could not scan file %q: %w", netdevName, err)
	}

	return recv, trns, nil
}

// Of course bytes.Fields allocates,
// so we use FieldsSeq to put the fields into dest.
func readFields(data []byte, dest [][]byte) (n int) {
	for field := range bytes.FieldsSeq(data) {
		if len(dest) == 0 {
			return
		}
		dest[0] = field
		dest = dest[1:]
		n++
	}
	return n
}
