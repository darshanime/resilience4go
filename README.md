# resilience4go


## Example


```go
package main

import "time"

func main() {
	bh := bulkhead.New("google.com").WithMaxParallelCalls(40).WithMaxWaitDuration(time.Second * 1)
	rt := retry.New().WithMaxRetries(20)
	newClient := resil.NewClient().WithClient(&myClient).WithBulkHead(bh).WithRetry(rt)
	newClient.Get("http://google.com")
}
```

This has more human friendly controls. Earlier, we could configure how many threads to use for an external http call. That is difficult to reason about, what does 2 threads mean in terms of cpu, memory?. It is difficult to control, will reducing threads from 2 to 1 reduce the number of parallel calls by half?
