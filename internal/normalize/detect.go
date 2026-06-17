package normalize

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

var (
	reURL    = regexp.MustCompile(`(?i)^https?://[^\s/$.?#].[^\s]*$`)
	reEmail  = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	reIPv6   = regexp.MustCompile(`(?i)^([0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}$`)
	reIPv4   = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	reMD5    = regexp.MustCompile(`(?i)^[a-f0-9]{32}$`)
	reSHA1   = regexp.MustCompile(`(?i)^[a-f0-9]{40}$`)
	reSHA256 = regexp.MustCompile(`(?i)^[a-f0-9]{64}$`)
	reSHA512 = regexp.MustCompile(`(?i)^[a-f0-9]{128}$`)
	reDomain = regexp.MustCompile(`(?i)^([a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)
)

func DetectType(raw string) (Type, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty value")
	}

	switch {
	case reURL.MatchString(raw):
		return TypeURL, nil

	case reEmail.MatchString(raw):
		return TypeEmail, nil

	case reIPv6.MatchString(raw):
		if _, err := netip.ParseAddr(raw); err == nil {
			return TypeIPv6, nil
		}

	case reIPv4.MatchString(raw):
		if addr, err := netip.ParseAddr(raw); err == nil && addr.Is4() {
			return TypeIP, nil
		}

	case reMD5.MatchString(raw):
		return TypeHash, nil
	case reSHA1.MatchString(raw):
		return TypeHash, nil
	case reSHA256.MatchString(raw):
		return TypeHash, nil
	case reSHA512.MatchString(raw):
		return TypeHash, nil

	case reDomain.MatchString(raw):
		return TypeDomain, nil
	}

	return "", fmt.Errorf("cannot detect type for value %q", raw)
}
