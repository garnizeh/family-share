package security

import (
	"net/netip"
	"sync"
)

var (
	trustedProxyMu     sync.RWMutex
	trustedProxyCIDRs  []netip.Prefix
)

// SetTrustedProxyCIDRs configures the trusted proxy CIDR ranges.
func SetTrustedProxyCIDRs(prefixes []netip.Prefix) {
	trustedProxyMu.Lock()
	trustedProxyCIDRs = append([]netip.Prefix(nil), prefixes...)
	trustedProxyMu.Unlock()
}

func getTrustedProxyCIDRs() []netip.Prefix {
	trustedProxyMu.RLock()
	defer trustedProxyMu.RUnlock()
	return append([]netip.Prefix(nil), trustedProxyCIDRs...)
}
