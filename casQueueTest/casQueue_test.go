package casQueueTest

import (
	"fmt"
	"github.com/golangCasQueue/casQueue"
	"sync"
	"testing"
	"time"
)

var goRoutineCnt = 1000
var batchSize = 32
var cacheSize = goRoutineCnt * 100

func checkPutAndGet(checkData []int){
	l := len(checkData)
	for i := 0; i < l; i++{
		if checkData[i] != i{
			fmt.Printf("\ncheckPutAndGet error!! lost data!!! i:%v checkData[i]:%v l:%v\n", i, checkData[i], l)
		}
	}
}

func BenchmarkChannel(b *testing.B) {
	ch := make(chan interface{}, cacheSize)
	var wg sync.WaitGroup
	wg.Add(1)

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch
		}
		wg.Done()
	}()

	for i := 0; i < b.N; i++ {
		ch <- i
	}
	wg.Wait()
}

func BenchmarkCasQueue(b *testing.B) {
	var checkData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wg sync.WaitGroup
	wg.Add(1)

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			val,ok,_:=q.Get()
			for !ok {
				val,ok,_=q.Get()
			}
			v:=val.(int)
			checkData[v] = v
		}
		wg.Done()
	}()

	for i := 0; i < b.N; i++ {
		ok, _ := q.Put(i)
		for !ok {
			ok, _ = q.Put(i)
		}
	}
	wg.Wait()
	checkPutAndGet(checkData)
}

func BenchmarkCasQueueWithBatch(b *testing.B) {
	var checkData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wg sync.WaitGroup
	wg.Add(1)

	b.ResetTimer()
	go func() {
		var valuesGet []interface{}
		valuesGet = make([]interface{}, batchSize)
		getSize := 0
		for i := 0; i < b.N; i+= getSize {
			getSize,_ = q.Gets(valuesGet)
			for j := 0; j < getSize; j++{
				v := valuesGet[j].(int)
				if v < b.N {
					checkData[v] = v
				}
			}
		}
		wg.Done()
	}()

	putSize := 0
	var valuesPut []interface{}
	valuesPut = make([]interface{}, batchSize)
	for i := 0; i < b.N; i+= putSize {
		for j:=0; j< batchSize; j++{
			valuesPut[j] = i + j
		}
		putSize, _ = q.Puts(valuesPut)
	}
	wg.Wait()
	checkPutAndGet(checkData)
}

func BenchmarkChannelReadContention(b *testing.B) {
	ch := make(chan interface{}, cacheSize)
	var wg sync.WaitGroup
	wg.Add(goRoutineCnt)
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			ch <- `a`
		}
	}()

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			for i := 0; i < b.N/goRoutineCnt; i++ {
				<-ch
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkCasQueueReadContention(b *testing.B) {
	var checkData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wgGet sync.WaitGroup
	wgGet.Add(goRoutineCnt)
	var wgPut sync.WaitGroup
	wgPut.Add(1)
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			ok, _ := q.Put(i)
			for !ok {
				ok, _ = q.Put(i)
			}
		}
		wgPut.Done()
	}()

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			for i := 0; i < b.N / goRoutineCnt; i++ {
				val, ok, _ := q.Get()
				for !ok {
					val, ok, _= q.Get()
				}
				v := val.(int)
				checkData[v] = v
			}
			wgGet.Done()
		}()
	}
	wgGet.Wait()
	wgPut.Wait()
	for q.Quantity() > 0{
		val, ok, _ := q.Get()
		for !ok {
			val, ok, _ = q.Get()
		}
		v := val.(int)
		checkData[v] = v
	}
	checkPutAndGet(checkData)
}

func BenchmarkCasQueueReadContentionWithBatch(b *testing.B) {
	var checkData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wgGet sync.WaitGroup
	wgGet.Add(goRoutineCnt)
	var wgPut sync.WaitGroup
	wgPut.Add(1)
	b.ResetTimer()

	go func() {
		putSize := 0
		var valuesPut []interface{}
		valuesPut = make([]interface{}, batchSize)
		for i := 0; i < b.N; i+= putSize {
			if i + batchSize > b.N{
				valuesPut = valuesPut[0:b.N-i]
			}
			l := len(valuesPut)
			for j:=0; j < l; j++{
				valuesPut[j] = i + j
			}
			putSize, _ = q.Puts(valuesPut)
		}
		wgPut.Done()
	}()

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			var valuesGet []interface{}
			valuesGet = make([]interface{}, batchSize)
			getSize := 0
			for i := 0; i < b.N / goRoutineCnt; i+= getSize {
				if i + batchSize > b.N / goRoutineCnt{
					valuesGet = make([]interface{}, b.N / goRoutineCnt - i)
				}
				getSize,_ = q.Gets(valuesGet)
				for j := 0; j < getSize; j++{
					v := valuesGet[j].(int)
					if v < b.N {
						checkData[v] = v
					}
				}
			}
			wgGet.Done()
		}()
	}
	wgGet.Wait()
	wgPut.Wait()
	for q.Quantity() > 0{
		val, ok, _ := q.Get()
		for !ok {
			val, ok, _ = q.Get()
		}
		v := val.(int)
		if v < b.N {
			checkData[v] = v
		}
	}
	checkPutAndGet(checkData)
}


func BenchmarkChannelContention(b *testing.B) {
	ch := make(chan interface{}, goRoutineCnt)
	var wg sync.WaitGroup
	wg.Add(goRoutineCnt * 2)
	b.ResetTimer()

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			for i := 0; i < b.N/goRoutineCnt; i++ {
				ch <- `a`
			}
			wg.Done()
		}()
	}

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			for i := 0; i < b.N/goRoutineCnt; i++ {
				<-ch
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkCasQueueContention(b *testing.B) {
	var checkData = make([]int, b.N)
	var putData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wgGet, wgPut sync.WaitGroup
	wgGet.Add(goRoutineCnt)
	wgPut.Add(goRoutineCnt)
	b.ResetTimer()

	for i := 0; i < goRoutineCnt; i++ {
		go func(i int, b *testing.B) {
			putNumber := 0
			for j := 0; j < b.N / goRoutineCnt; j++ {
				putNumber = i * (b.N / goRoutineCnt)+ j
				ok, _ := q.Put( putNumber)
				for !ok{
					ok, _ = q.Put(putNumber)
				}
				putData[putNumber]  = putNumber
			}
			wgPut.Done()
		}(i, b)
	}

	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			for i := 0; i < b.N / goRoutineCnt; i++ {
				val, ok, _:= q.Get()
				for !ok {
					val, ok, _= q.Get()
				}
				v:=val.(int)
				checkData[v] = v
			}
			wgGet.Done()
		}()
	}
	wgPut.Wait()
	l := len(putData)
	for i:= 0; i < l; i++{
		if putData[i] != i{
			ok, _ := q.Put(i)
			for ok == false {
				ok, _ = q.Put(i)
			}
		}
	}
	wgGet.Wait()
	for q.Quantity() > 0{
		val, ok, _ := q.Get()
		for !ok {
			val, ok, _ = q.Get()
		}
		v:=val.(int)
		checkData[v] = v
	}
	checkPutAndGet(checkData)
}

func BenchmarkCasQueueContentionWithBatch(b *testing.B) {
	var checkData = make([]int, b.N)
	var putData = make([]int, b.N)
	q := casQueue.NewQueue(uint64(cacheSize), time.Duration(10) * time.Microsecond)
	var wgGet, wgPut sync.WaitGroup
	wgGet.Add(goRoutineCnt)
	wgPut.Add(goRoutineCnt)
	b.ResetTimer()

	for i := 0; i < goRoutineCnt; i++ {
		go func(i int, b *testing.B) {
			putSize := 0
			var valuesPut []interface{}
			valuesPut = make([]interface{}, batchSize)
			putNumber := 0
			for j := 0; j < b.N / goRoutineCnt; j+=putSize {
				putNumber = i * (b.N / goRoutineCnt)+ j
				if putNumber + batchSize > (i+1) * (b.N / goRoutineCnt){
					valuesPut = valuesPut[0:(i+1) * (b.N / goRoutineCnt) - putNumber]
				}
				l := len(valuesPut)
				for k:=0; k < l; k++{
					valuesPut[k] = putNumber + k
				}
				putSize, _ = q.Puts(valuesPut)
				for k:=0; k < putSize; k++{
					v := valuesPut[k].(int)
					if v < b.N {
						putData[v] = v
					}
				}
			}
			wgPut.Done()
		}(i, b)
	}
	for i := 0; i < goRoutineCnt; i++ {
		go func() {
			var valuesGet []interface{}
			valuesGet = make([]interface{}, batchSize)
			getSize := 0
			for i := 0; i < b.N / goRoutineCnt; i+= getSize {
				if i + batchSize > b.N / goRoutineCnt{
					valuesGet = make([]interface{}, b.N / goRoutineCnt - i)
				}
				getSize,_ = q.Gets(valuesGet)
				for j := 0; j < getSize; j++{
					v := valuesGet[j].(int)
					if v < b.N {
						checkData[v] = v
					}
				}
			}
			wgGet.Done()
		}()
	}
	wgPut.Wait()
	l := len(putData)
	for i:= 0; i < l; i++{
		if putData[i] != i{
			ok, _ := q.Put(i)
			for ok == false {
				ok, _ = q.Put(i)
			}
		}
	}
	wgGet.Wait()
	for q.Quantity() > 0{
		val, ok, _ := q.Get()
		for !ok {
			val, ok, _ = q.Get()
		}
		v:=val.(int)
		if v < b.N {
			checkData[v] = v
		}
	}
	checkPutAndGet(checkData)
}
