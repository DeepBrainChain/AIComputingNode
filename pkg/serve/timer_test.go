package serve

import (
	"sync"
	"testing"
	"time"
)

func TestSingleTimer(t *testing.T) {
	timer := time.NewTimer(2 * time.Second)

	go func() {
		// select {
		// case <-timer.C:
		// 	t.Log("Timer triggered")
		// }
		<-timer.C
		t.Log("Timer triggered")
	}()

	time.Sleep(3 * time.Second)
}

func TestRepeatedTimer(t *testing.T) {
	timer := time.NewTimer(3 * time.Second)
	done := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Logf("%v | goroutine start", time.Now())
		for {
			select {
			case <-done:
				t.Logf("%v | goroutine over", time.Now())
				return
			case <-timer.C:
				t.Logf("%v | Timer triggered", time.Now())
			}

			time.Sleep(1 * time.Second)
			t.Logf("%v | Do something over", time.Now())
			timer.Reset(3 * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Second)
		done <- true
	}()

	wg.Wait()
	timer.Stop()
}

func TestTimerTicker(t *testing.T) {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				t.Log("goroutine over")
				return
			case cur := <-ticker.C:
				t.Log("tick at ", cur)
			}
		}
	}()

	time.Sleep(5 * time.Second)
	ticker.Stop()
	done <- true
}

func TestTimerTicker2(t *testing.T) {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			cur := <-ticker.C
			t.Log("tick at ", cur)
		}
	}()

	time.Sleep(5 * time.Second)
	ticker.Stop()
}

func TestGoroutine(t *testing.T) {
	done := make(chan bool)

	go func() {
		<-done
		t.Log("goroutine 1 over")
	}()

	go func() {
		<-done
		t.Log("goroutine 2 over")
	}()

	time.Sleep(1 * time.Second)
	done <- true
	done <- true
	time.Sleep(1 * time.Second)
}
