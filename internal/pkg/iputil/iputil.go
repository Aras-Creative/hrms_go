package iputil

import (
	"net"
	"strings"
	"sync"
)

var DefaultAllowedCIDRs []string

var (
	allowedNets   []*net.IPNet
	allowedNetsMu sync.RWMutex
)

// SetAllowedCIDRs replaces the global allow-list used by IsAllowed. Entries may
// be CIDR notation (e.g. "10.0.0.0/8") or a bare IP (e.g. "192.168.1.10"),
// which is normalized to a host route. Invalid entries are skipped. An empty
// list disables filtering, i.e. every IP is allowed.
func SetAllowedCIDRs(cidrs []string) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if ip := net.ParseIP(cidr); ip != nil {
				if ip.To4() != nil {
					cidr += "/32"
				} else {
					cidr += "/128"
				}
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		nets = append(nets, ipNet)
	}

	allowedNetsMu.Lock()
	allowedNets = nets
	DefaultAllowedCIDRs = append([]string(nil), cidrs...)
	allowedNetsMu.Unlock()
}

// IsAllowed reports whether ip belongs to the configured allow-list. When no
// CIDRs are configured every IP is allowed.
func IsAllowed(ip string) bool {
	allowedNetsMu.RLock()
	nets := allowedNets
	allowedNetsMu.RUnlock()

	if len(nets) == 0 {
		return true
	}

	host := ip
	if h, _, err := net.SplitHostPort(ip); err == nil {
		host = h
	}
	parsed := net.ParseIP(strings.TrimSpace(host))
	if parsed == nil {
		return false
	}

	for _, n := range nets {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}

// ValidateCIDR checks that cidr is a valid CIDR or a bare IP address.
func ValidateCIDR(cidr string) bool {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return false
	}
	if _, _, err := net.ParseCIDR(cidr); err == nil {
		return true
	}
	return net.ParseIP(cidr) != nil
}
