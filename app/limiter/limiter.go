package limiter

import (
	"sync"
	"time"
)

type GlobalLimiter struct {
	maxCount int
	count    int
	ticker   *time.Ticker
	mutex    sync.Mutex
}

func NewGlobalLimiter(maxCount int, d time.Duration) (*GlobalLimiter, func()) {
	l := &GlobalLimiter{
		maxCount: maxCount,
		count:    0,
		ticker:   time.NewTicker(d),
	}

	go func() {
		for {
			<-l.ticker.C
			l.mutex.Lock()
			l.count = 0
			l.mutex.Unlock()
		}
	}()

	return l, func() {
		l.ticker.Stop()
	}
}

func (l *GlobalLimiter) Allow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.count < l.maxCount {
		l.count++
		return true
	}

	return false
}
