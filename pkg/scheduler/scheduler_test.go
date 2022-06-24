package scheduler

import (
	"log"
	"os"
	"testing"
	"time"
)

func Clean() {
	os.RemoveAll("test")
}
func TestSingleTask(t *testing.T) {
	Clean()

	s, e := New("test")
	if e != nil {
		panic(e)
	}

	_, e = s.Add("SingleTask", time.Second*2)

	if e != nil {
		panic(e)
	}

	go s.Run()

	time.Sleep(time.Second * 5)

	s.Stop()
}

func TestDelayTask(t *testing.T) {
	Clean()

	s, e := New("test")
	if e != nil {
		panic(e)
	}

	go s.Run()

	s.Add("DelayTask", time.Second)

	time.Sleep(time.Second * 5)

	s.Stop()
}

func TestPeriodTask(t *testing.T) {
	Clean()

	s, e := New("test")
	if e != nil {
		panic(e)
	}

	go s.Run()

	task, e := s.Add("PeriodTask", time.Second*5, time.Second)

	if e != nil {
		panic(e)
	}

	time.Sleep(time.Second * 5)

	task.Stop()

	log.Println("stop task...")

	time.Sleep(time.Second * 5)

	s.Stop()
}
