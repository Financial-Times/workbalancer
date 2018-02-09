package workbalancer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var exists = struct{}{}

func TestBalancing_Ok(t *testing.T) {
	var workloads []Workload
	n := 32
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(4)

	actualResults := make(map[int]struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for result := range balancer.GetResults() {
			intResult, ok := result.(int)
			if !ok {
				t.Fatalf("A result was not of correct type. %v", result)
			}
			actualResults[intResult] = exists
		}
		wg.Done()
	}()
	balancer.Balance(workloads)
	wg.Wait()

	for i := 0; i < n; i++ {
		_, ok := actualResults[i+1]
		assert.True(t, ok, "expected work result %d wasn't found in results", i+1)
	}
}

func TestSingleWorker_Ok(t *testing.T) {
	var workloads []Workload
	n := 8
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(1)

	actualResults := make(map[int]struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for result := range balancer.GetResults() {
			intResult, ok := result.(int)
			if !ok {
				t.Fatalf("A result was not of correct type. %v", result)
			}
			// log.Infof("result=%v", intResult)
			actualResults[intResult] = exists
		}
		wg.Done()
	}()
	balancer.Balance(workloads)
	wg.Wait()

	for i := 0; i < n; i++ {
		_, ok := actualResults[i+1]
		assert.True(t, ok, "expected work result %d wasn't found in results", i+1)
	}
}

func TestNoWorker_Ok(t *testing.T) {
	var workloads []Workload
	n := 8
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(0)

	actualResults := make(map[int]struct{})
	wg := make(chan bool)
	go func() {
		for result := range balancer.GetResults() {
			t.Fatalf("A result appeared, but with no workers this should not have been possible. result=%v", result)
		}
		wg <- true
	}()
	timer := time.NewTimer(time.Second)
	select {
	case <-timer.C:
		return
	case <-wg:
		assert.Equal(t, actualResults, 0, "Result should be 0 because there are no workers.")
		balancer.Balance(workloads)
		t.Fatalf("Shouldn't end here as there are no workers to process work, hence no results.")

	}
}

// This basically tests that if there are two greater workloads, and a lot of small ones, and only two threads,
// then the total amount of time will be just the sum of the workloads and no threads will block waiting, not doing anything.
func TestTimeSumsWorkloadTimePerNumberOfWorkersEvenOnUnevenWorkUnits_Ok(t *testing.T) {
	var workloads []Workload
	n := 16
	u := 50
	workloads = append(workloads, &timeWork{t: time.Duration(n*u) * time.Millisecond})
	for i := 0; i < n-2; i++ {
		workloads = append(workloads, &timeWork{t: time.Duration(u) * time.Millisecond})
	}
	workloads = append(workloads, &timeWork{t: time.Duration(n*u) * time.Millisecond})
	balancer := NewChannelBalancer(2)

	start := time.Now()
	nActualResults := 0
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for _ = range balancer.GetResults() {
			nActualResults++
		}
		wg.Done()
	}()
	balancer.Balance(workloads)
	wg.Wait()

	expectedEnd := start.Add(time.Duration((2*n+1)*u) * time.Millisecond)
	assert.Equal(t, n, nActualResults)
	assert.True(t, time.Now().Before(expectedEnd))
}

type incWork struct {
	i int
}

func (w *incWork) Do() Result {
	return w.i + 1
}

type timeWork struct {
	t time.Duration
}

func (w *timeWork) Do() Result {
	time.Sleep(w.t)
	return true
}
