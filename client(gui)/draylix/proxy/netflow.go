package proxy

import (
	"draylix/drlx"
	"sync"
	"time"
)

var (
	AllTraffic   = int64(0)
	PerSecondOut = make(chan int, 32)
	OutData      = make(chan int, 512)
	PerSecondIn  = make(chan int, 32)
	InData       = make(chan int, 512)
	mu           = sync.Mutex{}
)

func StartCollectPerSecondData() {
	go perSecond(PerSecondOut, OutData)
	go perSecond(PerSecondIn, InData)
}

func perSecond(to, from chan int) {
	var p int
	ticker := time.NewTicker(time.Second)
	for {
		d := <-from
		select {
		case <-ticker.C:
			select {
			case to <- p:
				p = 0
			default:
				p = 0
			}
		default:
			p += d
		}
	}
}

func GetTraffic() string {
	mu.Lock()
	defer mu.Unlock()
	return drlx.BytesFormat(AllTraffic)
}

func TrafficIncrease(n int64) {
	mu.Lock()
	AllTraffic += n
	mu.Unlock()
}
