package anilist

import (
	"sync"
	"testing"
	"time"
)

func TestThrottle_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	th := newThrottle()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			th.wait()
			time.Sleep(1 * time.Millisecond)
			th.recordRateLimit()
			th.wait()
		}()
	}
	wg.Wait()
}
