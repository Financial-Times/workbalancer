package workbalancer

import (
	"sync"
)

type channelWorker struct {
	workloads       chan Workload
	workResults     chan<- WorkResult
	workerAvailable chan<- *channelWorker
	wg              *sync.WaitGroup
}

func newChannelWorker(workerAvailable chan<- *channelWorker, workResults chan<- WorkResult, wg *sync.WaitGroup) *channelWorker {
	worker := &channelWorker{
		workloads:       make(chan Workload, 1),
		workResults:     workResults,
		workerAvailable: workerAvailable,
		wg:              wg,
	}
	go worker.work()
	worker.workerAvailable <- worker
	return worker
}

func (w *channelWorker) addWork(workload Workload) {
	w.workloads <- workload
}

func (w *channelWorker) work() {
	for {
		workload, more := <-w.workloads
		// log.Infof("worker here, got workload=%v more=%v", workload, more)
		if !more {
			// log.Infof("worker calling w.wg.Done()")
			w.wg.Done()
			break
		}
		w.workResults <- workload.Do()
		// log.Infof("worker here, done with workload=%v", workload)
		w.workerAvailable <- w
	}
}

func (w *channelWorker) close() {
	close(w.workloads)
}
