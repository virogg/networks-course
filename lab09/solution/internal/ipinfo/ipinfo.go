package ipinfo

import (
	"fmt"
	"io"
	"net"
	"sort"
)

type entry struct {
	iface  string
	ip     net.IP
	mask   net.IPMask
	isIPv6 bool
}

func collect(includeAll bool) ([]entry, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("net.Interfaces: %w", err)
	}

	var out []entry
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if !includeAll && iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			isV6 := ipnet.IP.To4() == nil
			if !includeAll && isV6 {
				continue
			}
			out = append(out, entry{
				iface:  iface.Name,
				ip:     ipnet.IP,
				mask:   ipnet.Mask,
				isIPv6: isV6,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].iface != out[j].iface {
			return out[i].iface < out[j].iface
		}
		return out[i].ip.String() < out[j].ip.String()
	})
	return out, nil
}

func Print(w io.Writer, includeAll bool) error {
	entries, err := collect(includeAll)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(w, "no addresses found")
		return nil
	}
	for _, e := range entries {
		ones, _ := e.mask.Size()
		family := "ipv4"
		maskStr := net.IP(e.mask).String()
		if e.isIPv6 {
			family = "ipv6"
			maskStr = e.mask.String()
		}
		fmt.Fprintf(w, "%-12s %-5s ip=%s mask=%s /%d\n", e.iface, family, e.ip, maskStr, ones)
	}
	return nil
}
