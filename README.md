# workbalancer

Balances an array of workload into multiple goroutines.

For example if you have 100 operations to do, you can divide that to be done by 4 goroutines without fixing it to 25-25-25-25 static parts. The workers will take up work as they finish with current ones and none of the workers are waiting/blocking for the others to finish.

## Use:

```
    import "github.com/Financial-Times/workbalancer"

	balancer :=  workbalancer.NewChannelBalancer(parallelism)
	
	go func() {
		for result := range r.balancer.GetResults() {
			pResult, ok := result.(publishResult)
			if !ok {
				log.Errorf("Work result is not of expected type: %v", result)
			}
			for _, msg := range pResult.msgs {
				log.Info(msg)
				msgs = append(msgs, msg)
			}
			for _, err := range pResult.errs {
				log.Error(err)
				errs = append(errs, err)
			}
		}
		allResultsFetched.Done()
	}()

	var workloads []workbalancer.Workload
	for _, uuid := range uuids {
		workloads = append(workloads, &publishWork{
			uuid:            uuid,
			uuidRepublisher: r.uuidRepublisher,
			publishScope:    publishScope,
			tidPrefix:       tidPrefix,
		})
	}
	balancer.Balance(workloads)
	allResultsFetched.Wait()
	return msgs, errs

    func (w *exampleWork) Do() workbalancer.WorkResult {
	    msgs, errs := time.Sleep()
    }
	return publishResult{msgs, errs}

```