# casQueue
fork from https://github.com/yireyun/go-queue Fix it's bug && with full test

## usage
go get -u github.com/golangCasQueue/casQueue
```go
q := casQueue.NewQueue(1024*1024, time.Duration(10)* time.Microsecond)

// example data is int
func putExample(data int){
	// if queue is full, Put will sleep and runtime.Gosched()
	ok, _ := q.Put(data)
	for !ok {
		ok, _ = q.Put(data)
	}
}

// example data is int
func getExample() int {
	// if queue is empty, Get will sleep and runtime.Gosched()
	val, ok, _ := q.Get()
	for !ok {
		val, ok, _= q.Get()
	}
	return val.(int)
}

// batchPut sometime quicker
func putsExample(putData []int) {
	putSize := 0
	l := len(putData)
	for i := 0; i < l; i+= putSize {
		putSize, _ = q.Puts(putData)
		putData = putData[0..putSize]
	}
}

func getsExample(totalGetSize int) []interface{} {
	var valuesGet []interface{} = make([]interface{}, totalGetSize)
	var rtn [] interface{}
	getSize := 0
	for i := 0; i < totalGetSize; i+= getSize {
		getSize, _ = q.Gets(valueGet)
		rtn += valueGet[0..getSize]
	}
	return rtn
}
```

