# casQueue
fork from https://github.com/yireyun/go-queue Fix it's bug && with full test

## usage
go get -u github.com/golangCasQueue/casQueue
```go
q := casQueue.NewQueue(1024*1024, time.Duration(10)* time.Microsecond)

// example data is int
func putExample(data int){
	ok, _ := q.Put(data)
		for !ok {
			ok, _ = q.Put(data)
		}
}

// example data is int
func getExample() int {
	val, ok, _ := q.Get()
	for !ok {
		val, ok, _= q.Get()
	}
	return val.(int)
}
```

