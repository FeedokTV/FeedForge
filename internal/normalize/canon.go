package normalize

import (
	"net/netip"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

func CanonicalizeType(valueType Type, value string) string {
	switch valueType {
	case TypeURL:
		return canonicalURL(value)
	case TypeDomain:
		return canonicalDomain(value)
	case TypeIP:
		return canonicalIP(value)
	case TypeIPv6:
		return canonicalIP(value)
	case TypeHash:
		return canonicalHash(value)
	case TypeEmail:
		return canonicalEmail(value)
	default:
		return value
	}
}

func canonicalURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return strings.ToLower(raw)
	}

	u.Scheme = strings.ToLower(u.Scheme)

	host := strings.ToLower(u.Hostname())
	ascii, err := idna.ToASCII(host)
	if err == nil {
		host = ascii
	}

	port := u.Port()
	if (u.Scheme == "http" && port == "80") ||
		(u.Scheme == "https" && port == "443") {
		u.Host = host
	} else if port != "" {
		u.Host = host + ":" + port
	} else {
		u.Host = host
	}

	u.Fragment = ""

	return u.String()
}

func canonicalDomain(raw string) string {
	d := strings.ToLower(raw)

	d = strings.TrimSuffix(d, ".")

	ascii, err := idna.ToASCII(d)
	if err == nil {
		return ascii
	}

	return d
}

func canonicalIP(raw string) string {
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return raw
	}

	return addr.String()
}

func canonicalHash(raw string) string {
	return strings.ToLower(raw)
}

func canonicalEmail(raw string) string {
	parts := strings.SplitN(raw, "@", 2)
	if len(parts) != 2 {
		return raw
	}

	local := parts[0]
	domain := parts[1]

	domain = strings.ToLower(domain)

	ascii, err := idna.ToASCII(domain)
	if err == nil {
		domain = ascii
	}

	return local + "@" + domain
}
