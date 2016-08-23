package server

import (
	"testing"
	"time"
)

func TestDispatcher(t *testing.T) {
	jobQueue := make(chan Job)

	dispatcher := NewDispatcher(jobQueue, 1)
	dispatcher.Run()

	funcCalled := make(chan bool)

	jobQueue <- Job{func(req Request, res *Response) {
		funcCalled <- true
	}, Request{}, ResponseWriter{}}

	select {
	case <-funcCalled:
		return
	case <-time.After(2 * time.Second):
		t.Fatal("No job dispatched. Timedout.")
	}
}

func TestDispatcherNonIdle(t *testing.T) {
	jobQueue := make(chan Job)

	dispatcher := NewDispatcher(jobQueue, 1)
	dispatcher.Run()

	// get a jobChannel and only one is avaliable, so we expect a timeout.
	<-dispatcher.WorkerPool

	funcCalled := make(chan bool)

	jobQueue <- Job{func(req Request, res *Response) {
		funcCalled <- true
	}, Request{}, ResponseWriter{}}

	select {
	case <-funcCalled:
		t.Fatal("Job got called but no worker was idle")
	case <-time.After(2 * time.Second):
		return
	}
}

func TestDispatcherTwoWorkers(t *testing.T) {
	jobQueue := make(chan Job)

	dispatcher := NewDispatcher(jobQueue, 2)
	dispatcher.Run()

	// get a jobChannel but two are is avaliable
	<-dispatcher.WorkerPool

	funcCalled := make(chan bool)

	jobQueue <- Job{func(req Request, res *Response) {
		funcCalled <- true
	}, Request{}, ResponseWriter{}}

	select {
	case <-funcCalled:
		return
	case <-time.After(2 * time.Second):
		t.Fatal("No job dispatched. Timedout.")
	}
}
