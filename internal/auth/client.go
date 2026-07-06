// SPDX-License-Identifier: MIT
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
)

func ClientHash(pepper, ip, userAgent string) string {
	ip = NormalizeIP(ip)
	ua := strings.TrimSpace(userAgent)
	sum := sha256.Sum256([]byte(pepper + "|" + ip + "|" + ua))
	return hex.EncodeToString(sum[:])
}

func NormalizeIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	host := raw
	if h, _, err := net.SplitHostPort(raw); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return host
	}
	return ip.String()
}

func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			parts := strings.Split(fwd, ",")
			if len(parts) > 0 {
				if ip := strings.TrimSpace(parts[0]); ip != "" {
					return NormalizeIP(ip)
				}
			}
		}
		if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
			return NormalizeIP(real)
		}
	}
	return NormalizeIP(r.RemoteAddr)
}

func ClientUserAgent(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("User-Agent"))
}
