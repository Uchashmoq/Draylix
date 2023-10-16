package proxy

import (
	"draylix/drlx"
	"sync"
)

var (
	AllTraffic       = int64(0)
	PerSecondNetflow = make(chan int64, 32)
	NetflowChan      = make(chan int64, 1024)
	mu               = sync.Mutex{}
)

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
