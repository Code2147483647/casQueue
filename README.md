# casQueue
fork from https://github.com/yireyun/go-queue Fix it's bug && with strict test

## usage
go get -u github.com/golangCasQueue/casQueue
```go
package main

import (
	"fmt"
	"github.com/golangCasQueue/casQueue"
	"time"
)

var q = casQueue.NewQueue(1024*1024, time.Duration(10)* time.Microsecond)

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
func putsExample(putData []interface{}) {
	putSize := 0
	l := len(putData)
	for i := 0; i < l; i+= putSize {
		putSize, _ = q.Puts(putData)
		putData = putData[0:putSize]
	}
}

func getsExample(totalGetSize int) []interface{} {
	var valuesGet = make([]interface{}, totalGetSize)
	var rtn = make([]interface{}, 0)
	getSize := 0
	for i := 0; i < totalGetSize; i+= getSize {
		getSize, _ = q.Gets(valuesGet)
		rtn = append(rtn, valuesGet[0:getSize]...)
	}
	return rtn
}

func main() {
	putExample(1)
	fmt.Println(getExample())

	putsExample([]interface{}{2, 3, 4})
	fmt.Println(getsExample(3))
}
```

