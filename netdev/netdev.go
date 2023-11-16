package netdev

import(
	"os"
	"fmt"
	"strings"
	"strconv"
	"regexp"
)

func ReadNetDev() ([]byte, error) {
	netdev, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("could not read /proc/net/dev: %w", err)
	}
	return netdev, nil
}

func GetTraffic(netdev []byte) (int64, int64, error) {
	var recv int64 = 0
	var trns int64 = 0

	for i, line := range strings.Split(string(netdev), "\n") {
		if i < 2 || len(line) == 0 {
			continue
		}

		f := strings.Fields(line)
		iface       := f[0]
		recv_string := f[1]
		trns_string := f[9]

		// pattern for matching typical wan network interface names
		// https://www.thomas-krenn.com/en/wiki/Predictable_Network_Interface_Names
		// https://www.freedesktop.org/software/systemd/man/255/systemd.net-naming-scheme.html
		ptrn := `^(eth\d+|en[osp]\d+\S+|enx\S+|w[lw]\S+):$`

		match, err := regexp.MatchString(ptrn, iface)
		if err != nil {
			return 0, 0, fmt.Errorf("could not do regexp matching: %w", err)
		}
		if ! match {
			continue;
		}

		x, err := strconv.ParseInt(recv_string, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("could not convert string to int: %w", err)
		}
		recv += x

		x, err = strconv.ParseInt(trns_string, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("could not convert string to int: %w", err)
		}
		trns += x
	}

	return recv, trns, nil
}
