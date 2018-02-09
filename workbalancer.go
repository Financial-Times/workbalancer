package workbalancer

import (
	"sync"
)

type WorkBalancer interface {
	Balance(workloads []Workload)
	GetResults() <-chan Result
}

type channelBalancer struct {
	availableWorkers chan *worker
	results          chan Result
	workers          []*worker
	wg               *sync.WaitGroup
}

func NewChannelBalancer(nWorkers int) WorkBalancer {
	availableWorkers := make(chan *worker, nWorkers)
	results := make(chan Result, nWorkers)
	var workers []*worker
	wg := &sync.WaitGroup{}
	wg.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		worker := newWorker(availableWorkers, results, wg)
		workers = append(workers, worker)
	}
	return &channelBalancer{
		availableWorkers: availableWorkers,
		results:          results,
		workers:          workers,
		wg:               wg,
	}
}

type Workload interface {
	Do() Result
}

type Result interface {
}

func (b *channelBalancer) Balance(workloads []Workload) {
	for _, workload := range workloads {
		worker := <-b.availableWorkers
		worker.addWork(workload)
	}
	for _, worker := range b.workers {
		worker.close()
	}
	b.wg.Wait()
	close(b.results)
}

func (b *channelBalancer) GetResults() <-chan Result {
	return b.results
}
