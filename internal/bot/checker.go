package bot

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"
)

func (b *Bot) checkBaseURL(ctx context.Context, baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if !slices.Contains([]string{"http", "https"}, u.Scheme) {
		return fmt.Errorf("unsupported scheme '%s', must be http or https", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("host is required")
	}

	if u.User != nil {
		return fmt.Errorf("user info in URL is not allowed for security reasons")
	}

	host := u.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSpace(host))

	whitelist := b.profileGetStringArrayOr("security.baseapi.whitelist", nil)
	blacklist := b.profileGetStringArrayOr("security.baseapi.blacklist", nil)
	allowLocal := b.profileGetBoolOr("security.baseapi.allowLocal", false)

	if matchHostInList(host, blacklist) {
		return fmt.Errorf("base url host '%s' is blocked", host)
	}
	if len(whitelist) > 0 && !matchHostInList(host, whitelist) {
		return fmt.Errorf("base url host '%s' is not allowed", host)
	}

	if !allowLocal {
		if err := b.checkHostSafety(ctx, u.Host); err != nil {
			return fmt.Errorf("security check failed: %w", err)
		}
	}

	return nil
}

func matchHostInList(host string, list []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	for _, raw := range list {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		rawLower := strings.ToLower(raw)

		if strings.Contains(rawLower, "://") {
			if u, err := url.Parse(rawLower); err == nil {
				rawLower = strings.ToLower(strings.TrimSpace(u.Hostname()))
			}
		} else {
			rawLower = strings.TrimSuffix(rawLower, "/")
		}

		if rawLower == "" {
			continue
		}

		if rawLower == host {
			return true
		}
		if strings.HasPrefix(rawLower, "*.") {
			suffix := strings.TrimPrefix(rawLower, "*")
			if strings.HasSuffix(host, suffix) && host != strings.TrimPrefix(suffix, ".") {
				return true
			}
		}
		if strings.HasPrefix(rawLower, ".") && strings.HasSuffix(host, rawLower) && host != strings.TrimPrefix(rawLower, ".") {
			return true
		}
	}
	return false
}

func (b *Bot) checkHostSafety(ctx context.Context, host string) error {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		h = host
	}

	if ip := net.ParseIP(h); ip != nil {
		if b.checkIsPrivateIP(ip) {
			return fmt.Errorf("access to private IP '%s' is forbidden", h)
		}
		return nil
	}

	dnsCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIPAddr(dnsCtx, h)
	if err != nil {
		return fmt.Errorf("failed to resolve host '%s': %w", h, err)
	}

	for _, ipAddr := range ips {
		if b.checkIsPrivateIP(ipAddr.IP) {
			return fmt.Errorf("host '%s' resolves to private IP '%s', access forbidden", h, ipAddr.IP.String())
		}
	}

	return nil
}

func (b *Bot) checkIsPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return ip.IsPrivate()
}
