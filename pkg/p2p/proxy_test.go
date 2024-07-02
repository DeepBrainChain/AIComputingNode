package p2p

import (
	"fmt"
	"net/url"
	"testing"
)

func TestHttpUrl(t *testing.T) {
	var testUrl = "http://127.0.0.1:8080/api/v0/test?p1=hello&p2=100"

	u1, err := url.Parse(testUrl)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	t.Logf("Parse url success %s", u1.String())
	fmt.Println(*u1)

	u2, err := url.ParseRequestURI(testUrl)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	t.Logf("Parse url success %s", u2.String())
	fmt.Println(u2)

	_, err = url.Parse("")
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
}
