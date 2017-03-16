package porthos

import "github.com/porthos-rpc/porthos-go/log"

// Job represents the job to be run
type Job struct {
	Method         MethodHandler
	Request        Request
	ResponseWriter ResponseWriter
}

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

// NewWorker creates a new worker pool to run jobs.
func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

// Start method starts the run loop for the worker.
func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				res := newResponse()
				job.Method(job.Request, res)

				err := job.ResponseWriter.Write(res)

				if err != nil {
					log.Error("Error writing response: '%s'", err.Error())
				}
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
