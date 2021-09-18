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
