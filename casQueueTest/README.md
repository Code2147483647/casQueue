# casQueueTest
```
It's compare channel and casQueue(put,get) and casQueue(puts, gets)
It's a really strict test, It's will check the queue with all put and get data
```

## usage
```
go get -u github.com/golangCasQueue/casQueue
cd $GOPATH/src/github.com/golangCasQueue/casQueue/casQueueTest/
go test -bench="." -benchtime="10s"
```
## result
```go
BenchmarkChannel-4                          	100000000	       128 ns/op
BenchmarkEsQueue-4                          	200000000	       100 ns/op
BenchmarkEsQueueWithBatch-4                 	300000000	        46.4 ns/op
BenchmarkChannelReadContention-4            	50000000	       361 ns/op
BenchmarkEsQueueReadContention-4            	100000000	       115 ns/op
BenchmarkEsQueueReadContentionWithBatch-4   	300000000	        59.9 ns/op
BenchmarkChannelContention-4                	100000000	       175 ns/op
BenchmarkEsQueueContention-4                	100000000	       113 ns/op
BenchmarkEsQueueContentionWithBatch-4       	500000000	        83.2 ns/op
PASS
ok  	github.com/golangCasQueue/casQueue/casQueueTest	199.714s
```

