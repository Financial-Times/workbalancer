package workbalancer

import (
	"sync"
)

type worker struct {
	workloads        chan Workload
	results          chan<- Result
	availableWorkers chan<- *worker
	wg               *sync.WaitGroup
}

func newWorker(availableWorkers chan<- *worker, results chan<- Result, wg *sync.WaitGroup) *worker {
	worker := &worker{
		workloads:        make(chan Workload, 1),
		results:          results,
		availableWorkers: availableWorkers,
		wg:               wg,
	}
	go worker.work()
	worker.availableWorkers <- worker
	return worker
}

func (w *worker) addWork(workload Workload) {
	w.workloads <- workload
}

func (w *worker) work() {
	for workload := range w.workloads {
		w.results <- workload.Do()
		w.availableWorkers <- w
	}
	w.wg.Done()
}

func (w *worker) close() {
	close(w.workloads)
}
