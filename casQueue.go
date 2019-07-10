package casQueue

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

type casCache struct {
	putNo uint64
	getNo uint64
	value interface{}
}

// lock free queue
type CasQueue struct {
	sleepTime time.Duration
	capacity  uint64
	capMod    uint64
	putPos    uint64
	getPos    uint64
	cache     []casCache
}

func NewQueue(capacity uint64, sleepTime time.Duration) *CasQueue {
	q := new(CasQueue)
	q.capacity = minQuantity(capacity)
	q.capMod = q.capacity - 1
	q.putPos = 0
	q.getPos = 0
	q.sleepTime = sleepTime
	q.cache = make([]casCache, q.capacity)
	for i := range q.cache {
		cache := &q.cache[i]
		cache.getNo = uint64(i)
		cache.putNo = uint64(i)
	}
	cache := &q.cache[0]
	cache.getNo = q.capacity
	cache.putNo = q.capacity
	return q
}

func (q *CasQueue) String() string {
	getPos := atomic.LoadUint64(&q.getPos)
	putPos := atomic.LoadUint64(&q.putPos)
	return fmt.Sprintf("Queue{capacity: %v, capMod: %v, putPos: %v, getPos: %v}",
		q.capacity, q.capMod, putPos, getPos)
}

func (q *CasQueue) Capacity() uint64 {
	return q.capacity
}

func (q *CasQueue) Quantity() uint64 {
	var putPos, getPos uint64
	var quantity uint64
	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos > getPos {
		quantity = putPos - getPos
	} else {
		quantity = 0
	}
	return quantity
}

// put queue functions
func (q *CasQueue) Put(val interface{}) (ok bool, quantity uint64) {
	var putPos, putPosNew, getPos, posCnt uint64
	var cache *casCache

	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos > getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = 0
	}

	// full
	if posCnt >= q.capacity {
		time.Sleep(q.sleepTime)
		return false, posCnt
	}

	putPosNew = putPos + 1
	if !atomic.CompareAndSwapUint64(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return false, posCnt
	}

	// putPosNew&q.capMod == putPosNew % q.capacity when q.capacity is 2 ^ n
	cache = &q.cache[putPosNew&q.capMod]

	for {
		getNo := atomic.LoadUint64(&cache.getNo)
		putNo := atomic.LoadUint64(&cache.putNo)
		if putPosNew == putNo && getNo == putNo {
			cache.value = val
			atomic.AddUint64(&cache.putNo, q.capacity)
			return true, posCnt + 1
		} else {
			runtime.Gosched()
		}
	}
}

// get queue functions
func (q *CasQueue) Get() (val interface{}, ok bool, quantity uint64) {
	var putPos, getPos, getPosNew, posCnt uint64
	var cache *casCache

	putPos = atomic.LoadUint64(&q.putPos)
	getPos = atomic.LoadUint64(&q.getPos)

	if putPos > getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = 0
	}

	if posCnt < 1 {
		time.Sleep(q.sleepTime)
		return nil, false, posCnt
	}

	getPosNew = getPos + 1
	if !atomic.CompareAndSwapUint64(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return nil, false, posCnt
	}

	// putPosNew&q.capMod == putPosNew % q.capacity when q.capacity is 2 ^ n
	cache = &q.cache[getPosNew&q.capMod]

	for {
		getNo := atomic.LoadUint64(&cache.getNo)
		putNo := atomic.LoadUint64(&cache.putNo)
		if getPosNew == getNo && getNo == putNo-q.capacity {
			val = cache.value
			cache.value = nil
			atomic.AddUint64(&cache.getNo, q.capacity)
			return val, true, posCnt - 1
		} else {
			runtime.Gosched()
		}
	}
}

// puts queue functions
func (q *CasQueue) Puts(values []interface{}) (puts, quantity int) {
	var putPos, putPosNew, getPos, posCnt, putCnt uint64

	getPos = atomic.LoadUint64(&q.getPos)
	putPos = atomic.LoadUint64(&q.putPos)

	if putPos > getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = 0
	}

	if posCnt >= q.capacity {
		time.Sleep(q.sleepTime)
		return 0, int(posCnt)
	}

	if capPuts, size := q.capacity-posCnt, uint64(len(values)); capPuts >= size {
		putCnt = size
	} else {
		putCnt = capPuts
	}
	putPosNew = putPos + putCnt

	if !atomic.CompareAndSwapUint64(&q.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return 0, int(posCnt)
	}

	for posNew, v := putPos+1, uint64(0); v < putCnt; posNew, v = posNew+1, v+1 {
		// putPosNew&q.capMod == putPosNew % q.capacity when q.capacity is 2 ^ n
		var cache = &q.cache[posNew&q.capMod]
		for {
			getNo := atomic.LoadUint64(&cache.getNo)
			putNo := atomic.LoadUint64(&cache.putNo)
			if posNew == putNo && getNo == putNo {
				cache.value = values[v]
				atomic.AddUint64(&cache.putNo, q.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return int(putCnt), int(posCnt + putCnt)
}

// gets queue functions
func (q *CasQueue) Gets(values []interface{}) (gets, quantity int) {
	var putPos, getPos, getPosNew, posCnt, getCnt uint64

	putPos = atomic.LoadUint64(&q.putPos)
	getPos = atomic.LoadUint64(&q.getPos)

	if putPos > getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = 0
	}

	if posCnt < 1 {
		time.Sleep(q.sleepTime)
		return 0, int(posCnt)
	}

	if size := uint64(len(values)); posCnt >= size {
		getCnt = size
	} else {
		getCnt = posCnt
	}
	getPosNew = getPos + getCnt

	if !atomic.CompareAndSwapUint64(&q.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return 0, int(posCnt)
	}

	for posNew, v := getPos+1, uint64(0); v < getCnt; posNew, v = posNew+1, v+1 {
		// putPosNew&q.capMod == putPosNew % q.capacity when q.capacity is 2 ^ n
		var cache = &q.cache[posNew&q.capMod]
		for {
			getNo := atomic.LoadUint64(&cache.getNo)
			putNo := atomic.LoadUint64(&cache.putNo)
			if posNew == getNo && getNo == putNo-q.capacity {
				values[v] = cache.value
				cache.value = nil
				getNo = atomic.AddUint64(&cache.getNo, q.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}

	return int(getCnt), int(posCnt - getCnt)
}

// round 到最近的2的倍数
func minQuantity(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
