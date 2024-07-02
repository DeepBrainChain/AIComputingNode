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
	fmt.Println(u1)

	u2, err := url.ParseRequestURI(testUrl)
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
	t.Logf("Parse url success %s", u2.String())
	t.Logf("Query list %v", u2.Query())
	queryValues := u2.Query()
	queryValues.Add("project", "a b")
	queryValues.Add("model", "p2p")
	u2.RawQuery = queryValues.Encode()
	t.Logf("Modify url querys success %s", u2.String())
	newQuery := u2.Query()
	t.Logf("Parse query project %s model %s", newQuery.Get("project"), newQuery.Get("model"))

	_, err = url.Parse("")
	if err != nil {
		t.Fatalf("Parse url failed: %v", err)
	}
}
