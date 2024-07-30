package serve

import (
	"net/http"
	"testing"
)

func TestNewHttpRequest(t *testing.T) {
	req, err := http.NewRequest("POST", "http://192.168.1.159:7001/api/v0/chat/completion", nil)
	if err != nil {
		t.Fatalf("http.NewRequest failed %v", err)
	}
	nr := new(http.Request)
	*nr = *req
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nr.URL.String())
	nr.URL.Scheme = "https"
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nr.URL.String())
	nr.URL.Host = "ai.dbc.org"
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nr.URL.String())

	nrt := req.Clone(req.Context())
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nrt.URL.String())
	nrt.URL.Scheme = "http"
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nrt.URL.String())
	nrt.URL.Host = "192.168.1.159:7001"
	t.Logf("origin request url %v", req.URL.String())
	t.Logf("new request url %v", nrt.URL.String())
}
