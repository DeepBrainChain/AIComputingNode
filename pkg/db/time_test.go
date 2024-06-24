package db

import (
	"testing"
	"time"
)

func TestTimeDiff(t *testing.T) {
	t1 := time.Now()
	t.Log("current time", t1)
	t1i64 := t1.Unix()
	t.Log("current time in unix", t1i64)
	t2 := time.Unix(t1i64, 0)
	t.Log("time from int64", t2)
	td1 := t2.Sub(t1)
	t.Log("time sub", td1)

	st := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local)
	t.Log("start time", st)
	td2 := t2.Sub(st)
	t.Log("time sub", td2)
	t.Logf("time diff %.2f days ~ %.f hours ~ %.f minutes ~ %.f seconds",
		td2.Hours()/24, td2.Hours(), td2.Minutes(), td2.Seconds())
}

func TestTimeAdd(t *testing.T) {
	t1 := time.Now()
	t.Log("current time", t1)
	t2 := t1.Add(time.Hour * 24)
	t.Log("add hours", t2)
	t.Log("new time after now", t2.After(t1))
	t3 := t1.Add(-time.Hour * 24)
	t.Log("add -hours", t3)
	t.Log("new time before now", t3.Before(t1))
}
