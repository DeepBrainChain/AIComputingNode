package timer

import (
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// go test -v -timeout 30s -count=1 -run TestGoCron AIComputingNode/pkg/timer
func TestGoCron(t *testing.T) {
	s, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("gocron.NewScheduler failed: %v", err)
	}

	job, err := s.NewJob(
		gocron.DurationJob(3*time.Second),
		gocron.NewTask(
			func(a string, b int) {
				t.Logf("%v %v", a, b)
			},
			"hello",
			1,
		),
	)
	if err != nil {
		t.Fatalf("NewJob error: %v", err)
	}

	t.Log(job.ID())

	s.Start()

	<-time.After(20 * time.Second)

	if err := s.Shutdown(); err != nil {
		t.Fatalf("gocron.Shutdown error: %v", err)
	}
	t.Log("gocron stoped")
}
