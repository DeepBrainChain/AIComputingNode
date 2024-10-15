package types

import (
	"net"
	"net/url"
	"testing"
)

func TestLocalAddrs(t *testing.T) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("InterfaceAddrs failed: %v", err)
	}
	for i, addr := range addrs {
		t.Logf("InterfaceAddrs[%v] %v %v", i, addr.Network(), addr.String())
	}

	purl, err := url.Parse("http://127.0.0.1:8080/api/model")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Parse host from url", purl.Hostname())

	ip := net.ParseIP(purl.Hostname())
	if ip == nil {
		t.Fatal("Parse ip from url failed")
	}
	for _, addr := range addrs {
		var ipNet *net.IPNet
		var ok bool
		if ipNet, ok = addr.(*net.IPNet); !ok {
			continue
		}

		if ipNet.IP.Equal(ip) {
			t.Log("ip matched !!!")
			break
		}
	}

	as, err := net.LookupHost("localhost")
	if err != nil {
		t.Fatalf("LookupHost failed: %v", err)
	}
	t.Log("LookupHost", as)

	t.Log("ParseIP localhost", net.ParseIP("localhost"))
	ip, ipnet, err := net.ParseCIDR("localhost")
	t.Log("ParseIP localhost", ip, ipnet, err)
}
