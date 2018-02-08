# workbalancer

Balances an array of workload into multiple goroutines.

For example if you have 100 operations to do, you can divide that to be done by 4 goroutines without fixing it to 25-25-25-25 static parts. The workers will take up work as they finish with current ones and none of the workers are waiting/blocking for the others to finish.

## Use:

```
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Financial-Times/workbalancer"
)

// A workload, in our example, consists of sleeping and returning how much we did.
type sleepyWork struct {
	t time.Duration
}

func (w *sleepyWork) Do() workbalancer.WorkResult {
	time.Sleep(w.t)
	return w.t
}

func main() {
	balancer := workbalancer.NewChannelBalancer(4)

	// to wait for all the results
	wg := sync.WaitGroup{}
	wg.Add(1)

	// need to continuously take from the results channel
	go func() {
		for result := range balancer.GetResults() {
			fmt.Printf("slept %v\n", result.(time.Duration).Nanoseconds()/1000000)
		}
		wg.Done()
	}()

	// create the workload
	var works []workbalancer.Workload
	for i := 0; i < 32; i++ {
		works = append(works, &sleepyWork{time.Duration(rand.Intn(200)) * time.Millisecond})
	}

	// work it!
	balancer.Balance(works)

	// wait for the prints to happen before exiting the game
	wg.Wait()
}
```

## Elaboration

Say you wanted to divide workloads coming from a channel into two worker channels, which can run and consume concurrently:

```
func selectFromTwo(mainChannel <-chan int, worker1, worker2 chan<- int) {
	for work := range mainChannel {
		select {
		case worker1 <- work:
		case worker2 <- work:
		}
	}
}
```

This works but without blocking unless both of them are busy, but that's all right until all the workers can run and none of them wait.

But how do you solve this for n workers?

```
func selectFromN(mainChannel <-chan int, workers []chan<- int) {
	for work := range mainChannel {
		for i := 0; ; i++ {
			workers[i%len(workers)] <- work
		}
	}
}
```

This would block on the first worker that is in the middle of a job, but there might be other workers in the array that might be empty.

This project resolves this problem by proactively pulling jobs to workers when they are done with a bit of more communication.

It might be that this can be solved in a simpler way, but this is what I've found for now.
