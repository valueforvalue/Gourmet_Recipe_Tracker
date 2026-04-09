package main

import "testing"

func TestIsSafeURL(t *testing.T) {
	cases := []struct {
		url   string
		safe  bool
		label string
	}{
		// Use public IP literals for "safe" tests to avoid DNS dependency in CI.
		{"https://8.8.8.8/api", true, "public IP (Google DNS)"},
		{"http://1.1.1.1/api", true, "public IP (Cloudflare)"},
		{"http://localhost/admin", false, "localhost"},
		{"http://127.0.0.1/secret", false, "loopback IPv4"},
		{"http://[::1]/secret", false, "loopback IPv6"},
		{"http://192.168.1.1/router", false, "private RFC1918 (192.168.x.x)"},
		{"http://10.0.0.1/internal", false, "private RFC1918 (10.x.x.x)"},
		{"http://172.16.0.1/internal", false, "private RFC1918 (172.16.x.x)"},
		{"http://169.254.169.254/latest/meta-data/", false, "link-local (AWS metadata)"},
		{"ftp://example.com/file", false, "non-http/https scheme"},
		{"file:///etc/passwd", false, "file scheme"},
		{"", false, "empty string"},
		{"not-a-url", false, "invalid URL"},
		{"http://this-hostname-does-not-exist.invalid/path", false, "unresolvable hostname"},
	}

	for _, c := range cases {
		got := isSafeURL(c.url)
		if got != c.safe {
			t.Errorf("[%s] isSafeURL(%q) = %v, want %v", c.label, c.url, got, c.safe)
		}
	}
}
