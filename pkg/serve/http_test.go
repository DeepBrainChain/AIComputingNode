package serve

import (
	"net/http"
	"sync"
	"testing"
	"time"
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

// https://juejin.cn/post/7033671944587182087
// Incorrect examples will cause panic
// go test -v -timeout 30s -count=1 -run TestChannel1 AIComputingNode/pkg/serve
func TestChannel1(t *testing.T) {
	ch := make(chan int)
	wg := sync.WaitGroup{}

	closed := false
	mutex := sync.Mutex{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)
		mutex.Lock()
		if !closed {
			ch <- 100
		}
		mutex.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case value, ok := <-ch:
			t.Logf("read channel %v %v", value, ok)
		case <-time.After(3 * time.Second):
			t.Log("timeout")
		}
		mutex.Lock()
		if !closed {
			close(ch)
			closed = true
		}
		mutex.Unlock()
	}()

	wg.Wait()
}

// go test -v -timeout 30s -count=1 -run TestChannel1 AIComputingNode/pkg/serve
func TestChannel2(t *testing.T) {
	ch := make(chan int)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)
		defer func() {
			if r := recover(); r != nil {
				t.Log("Recovered from panic:", r)
			}
		}()
		// 使用 select 非阻塞地发送数据，如果通道已经关闭，跳过发送
		select {
		case ch <- 100:
			t.Log("sent 100 to channel")
		default:
			t.Log("channel closed, skipping send")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case value, ok := <-ch:
			t.Logf("read channel %v %v", value, ok)
		case <-time.After(3 * time.Second):
			t.Log("timeout")
		}
		// 确保通道只关闭一次
		close(ch)
	}()

	wg.Wait()
}

// go test -v -timeout 30s -count=1 -run TestChannel3 AIComputingNode/pkg/serve
func TestChannel3(t *testing.T) {
	// ch := make(chan int)
	ch := make(chan int, 1)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)
		ch <- 100
		close(ch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case value, ok := <-ch:
			if ok {
				t.Logf("read channel %v", value)
			} else {
				t.Log("channel closed")
			}
		case <-time.After(3 * time.Second):
			t.Log("timeout")
		}
	}()

	wg.Wait()
}
