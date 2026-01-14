package middleware

import (
	"net/http"
	"strings"
)

// PublicBaseURL returns the public-facing base URL for the incoming request. It
// prefers proxy headers (X-Forwarded-Host/X-Forwarded-Proto, or Forwarded) and
// falls back to request Host/TLS. If it cannot determine a host, it returns
// fallbackBaseURL.
//
// The returned value will look like: "http(s)://example.com[:port]".
func PublicBaseURL(r *http.Request, fallbackBaseURL string) string {
	host := firstForwardedHost(r)
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}

	// If we still don't have a host, fall back to configured base URL.
	if host == "" {
		return strings.TrimRight(strings.TrimSpace(fallbackBaseURL), "/")
	}

	proto := firstForwardedProto(r)
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	return strings.TrimRight(proto, " :/") + "://" + host
}

func firstForwardedHost(r *http.Request) string {
	// Prefer X-Forwarded-Host (can be a comma-separated list).
	if xfh := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); xfh != "" {
		// Many proxies append their value to any existing one; taking the last
		// entry is safer if a client tries to spoof an earlier value.
		return lastCSVValue(xfh)
	}

	// RFC 7239 Forwarded: for=...,proto=...,host=...
	if fwd := strings.TrimSpace(r.Header.Get("Forwarded")); fwd != "" {
		// Many proxies append entries; take the last entry.
		last := lastCSVValue(fwd)
		parts := strings.Split(last, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "host=") {
				val := strings.TrimSpace(p[5:])
				return strings.Trim(val, `"`)
			}
		}
	}

	return ""
}

func firstForwardedProto(r *http.Request) string {
	// Prefer X-Forwarded-Proto (can be comma-separated).
	if xfp := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xfp != "" {
		// Many proxies append their value to any existing one; taking the last
		// entry is safer if a client tries to spoof an earlier value.
		return lastCSVValue(xfp)
	}

	// RFC 7239 Forwarded: for=...,proto=...,host=...
	if fwd := strings.TrimSpace(r.Header.Get("Forwarded")); fwd != "" {
		last := lastCSVValue(fwd)
		parts := strings.Split(last, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "proto=") {
				val := strings.TrimSpace(p[6:])
				return strings.Trim(val, `"`)
			}
		}
	}

	return ""
}

func lastCSVValue(v string) string {
	parts := strings.Split(v, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.TrimSpace(parts[i])
		if p != "" {
			return p
		}
	}
	return ""
}
