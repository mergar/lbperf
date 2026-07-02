package clientip

import (
	"net"
	"net/http"
	"strings"
)

func FromRequest(r *http.Request, trustedAddrs []string) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}

	if isTrusted(remoteHost, trustedAddrs) {
		if ip := parseForwardedFor(r.Header.Get("X-Forwarded-For")); ip != "" {
			return ip
		}
	}

	if remoteHost == "" {
		return "0.0.0.0"
	}
	return remoteHost
}

func isTrusted(addr string, trusted []string) bool {
	for _, t := range trusted {
		if addr == t {
			return true
		}
	}
	return false
}

func parseForwardedFor(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		ip := strings.TrimSpace(part)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}
