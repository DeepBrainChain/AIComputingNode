package ps

import (
	"fmt"
	"strings"
	"testing"
)

func TestHttpUrl2Multiaddr(t *testing.T) {
	var url string = "http://192.168.1.159:4002"
	urlParts := strings.Split(url, "//")
	t.Log(len(urlParts))
	t.Log(urlParts)
	if len(urlParts) != 2 {
		t.Fatal("empty array when split using //")
	}
	url = urlParts[1]

	urlParts = strings.Split(url, "/")
	t.Log(urlParts)
	url = urlParts[0]

	urlParts = strings.Split(url, ":")
	if len(urlParts) != 2 {
		t.Fatal("can not get ip and port")
	}
	t.Log(urlParts[0], urlParts[1])
	addr := fmt.Sprintf("/ip4/%s/tcp/%s", urlParts[0], urlParts[1])
	t.Log(addr)
	t.Log(HttpUrl2Multiaddr("http://192.168.1.159:4002"))
}
