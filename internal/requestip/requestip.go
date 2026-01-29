package requestip

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// ParseTrustedProxyCIDRs parses a comma-separated list of CIDR ranges.
func ParseTrustedProxyCIDRs(value string) ([]netip.Prefix, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", trimmed, err)
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}

// ClientIP returns the validated client IP for a request.
// Forwarded headers are only honored when the request originated
// from a trusted proxy CIDR range.
func ClientIP(r *http.Request, trustedProxyCIDRs []netip.Prefix) string {
	remoteIP := remoteAddrIP(r.RemoteAddr)
	if !isTrustedProxy(remoteIP, trustedProxyCIDRs) {
		return remoteIP
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			if addr, ok := parseAddr(part); ok {
				return addr.String()
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if addr, ok := parseAddr(xri); ok {
			return addr.String()
		}
	}

	return remoteIP
}

func isTrustedProxy(remoteIP string, trustedProxyCIDRs []netip.Prefix) bool {
	if len(trustedProxyCIDRs) == 0 {
		return false
	}

	addr, err := netip.ParseAddr(remoteIP)
	if err != nil {
		return false
	}

	for _, prefix := range trustedProxyCIDRs {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

func remoteAddrIP(remoteAddr string) string {
	if addr, ok := parseAddr(remoteAddr); ok {
		return addr.String()
	}
	return strings.TrimSpace(remoteAddr)
}

func parseAddr(value string) (netip.Addr, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return netip.Addr{}, false
	}

	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		trimmed = host
	}

	trimmed = strings.Trim(trimmed, "[]")
	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr, true
}
