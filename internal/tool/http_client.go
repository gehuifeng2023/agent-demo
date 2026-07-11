package tool

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxHTTPResponseBytes int64 = 1 << 20

type HTTPClient struct {
	client     *http.Client
	allowHosts map[string]struct{}
}

func NewHTTPClient(allowHosts []string, timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	allowed := make(map[string]struct{}, len(allowHosts))
	for _, host := range allowHosts {
		host = normalizeHTTPHost(host)
		if host != "" {
			allowed[host] = struct{}{}
		}
	}

	client := &HTTPClient{allowHosts: allowed}
	client.client = &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return client.allowedURL(req.URL)
		},
	}
	return client
}

func (c *HTTPClient) allowed(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}
	return c.allowedURL(u)
}

func (c *HTTPClient) allowedURL(u *url.URL) error {
	if u == nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must be absolute")
	}
	if scheme := strings.ToLower(u.Scheme); scheme != "http" && scheme != "https" {
		return fmt.Errorf("URL scheme not allowed: %s", u.Scheme)
	}

	host := normalizeHTTPHost(u.Hostname())
	if host == "" {
		return fmt.Errorf("URL host is empty")
	}
	if _, ok := c.allowHosts[host]; !ok {
		return fmt.Errorf("host not allowed: %s", host)
	}
	return nil
}

func normalizeHTTPHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}
