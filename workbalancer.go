package workbalancer

import (
	"sync"
)

type Workbalancer interface {
	Balance(workloads []Workload)
	GetResults() <-chan WorkResult
}

type channelBalancer struct {
	workerAvailable chan *channelWorker
	workResults     chan WorkResult
	workers         []*channelWorker
	wg              *sync.WaitGroup
}

func NewChannelBalancer(nWorkers int) Workbalancer {
	workerAvailable := make(chan *channelWorker, nWorkers)
	workResults := make(chan WorkResult, nWorkers)
	workers := []*channelWorker{}
	wg := &sync.WaitGroup{}
	wg.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		worker := newChannelWorker(workerAvailable, workResults, wg)
		workers = append(workers, worker)
	}
	return &channelBalancer{
		workerAvailable: workerAvailable,
		workResults:     workResults,
		workers:         workers,
		wg:              wg,
	}
}

type Workload interface {
	Do() WorkResult
}

type WorkResult interface {
}

func (b *channelBalancer) Balance(workloads []Workload) {
	for _, workload := range workloads {
		// log.Infof("at workload %v", workload)
		worker := <-b.workerAvailable
		// log.Infof("worker became available")
		worker.addWork(workload)
	}
	for _, worker := range b.workers {
		worker.close()
	}
	b.wg.Wait()
	close(b.workResults)
}

func (b *channelBalancer) GetResults() <-chan WorkResult {
	return b.workResults
}
