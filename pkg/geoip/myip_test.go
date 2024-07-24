package geoip

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// https://github.com/fufuok/myip/blob/master/myip.go
func TestGetPublicIP(t *testing.T) {
	ipv4 := []string{
		"https://4.ipw.cn",
		"https://api-ipv4.ip.sb/ip",
		"https://httpbin.org/ip",
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://ident.me",
		"https://myexternalip.com/raw",
		// "http://ip.42.pl/short",
		"https://ipinfo.io/ip",
	}
	// "ipv6": {
	// 	"https://6.ipw.cn",
	// 	"https://api-ipv6.ip.sb/ip",
	// 	"https://api64.ipify.org",
	// 	"http://ifconfig.me/ip",
	// 	"http://ident.me",
	// },
	// proxyUrl, err := url.Parse("http://127.0.0.1:10809")
	// if err != nil {
	// 	t.Fatalf("Parse proxy url failed %v", err)
	// }
	var DefaultTransport http.RoundTripper = &http.Transport{
		// Proxy: http.ProxyURL(proxyUrl),
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		DisableCompression: true,
		// ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		DisableKeepAlives:     false,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       128,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	for _, api := range ipv4 {
		req, err := http.NewRequest("GET", api, nil)
		if err != nil {
			t.Errorf("New http request failed %v", err)
			continue
		}
		resp, err := DefaultTransport.RoundTrip(req)
		if err != nil {
			t.Errorf("Round trip failed %v", err)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Read response failed %v", err)
			continue
		}
		t.Logf("%s -> %s", api, string(body))
	}
}
