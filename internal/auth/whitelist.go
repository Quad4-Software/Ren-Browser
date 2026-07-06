// SPDX-License-Identifier: MIT
package auth

import (
	"net"
	"strings"
)

type Whitelist struct {
	ips  map[string]struct{}
	nets []*net.IPNet
}

func ParseWhitelist(raw string) (*Whitelist, error) {
	w := &Whitelist{ips: make(map[string]struct{})}
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "/") {
			_, network, err := net.ParseCIDR(part)
			if err != nil {
				return nil, err
			}
			w.nets = append(w.nets, network)
			continue
		}
		ip := net.ParseIP(part)
		if ip == nil {
			return nil, &net.AddrError{Err: "invalid IP or CIDR", Addr: part}
		}
		w.ips[ip.String()] = struct{}{}
	}
	return w, nil
}

func (w *Whitelist) Allows(ip string) bool {
	if w == nil {
		return false
	}
	ip = NormalizeIP(ip)
	if ip == "" {
		return false
	}
	if _, ok := w.ips[ip]; ok {
		return true
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, network := range w.nets {
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}
