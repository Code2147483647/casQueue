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
go test -bench="." -benchtime="10s"
goos: darwin
goarch: amd64
pkg: github.com/golangCasQueue/casQueue/casQueueTest
BenchmarkChannel-4                           	100000000	       120 ns/op
BenchmarkCasQueue-4                          	200000000	        95.8 ns/op
BenchmarkCasQueueWithBatch-4                 	500000000	        54.4 ns/op
BenchmarkChannelReadContention-4             	50000000	       385 ns/op
BenchmarkCasQueueReadContention-4            	100000000	       112 ns/op
BenchmarkCasQueueReadContentionWithBatch-4   	300000000	        65.4 ns/op
BenchmarkChannelContention-4                 	100000000	       177 ns/op
BenchmarkCasQueueContention-4                	100000000	       109 ns/op
BenchmarkCasQueueContentionWithBatch-4       	500000000	        73.8 ns/op
PASS
ok  	github.com/golangCasQueue/casQueue/casQueueTest	203.069s
```

