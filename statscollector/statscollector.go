package main

import (
	"errors"
	"sync"
)

var (
	ErrRecordMoreThanTimeout = errors.New("record more that timeout")
)

type StatsCollector struct {
	mu      sync.RWMutex
	sum     int64
	count   int64
	timings []int64
	timeout int64
}

func NewStatsCollector(timeout int64) *StatsCollector {
	return &StatsCollector{
		timings: make([]int64, timeout+1),
		timeout: timeout,
	}
}

func (st *StatsCollector) Record(responseTimeMs ...int64) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, r := range responseTimeMs {
		if r > st.timeout {
			return ErrRecordMoreThanTimeout
		}
		st.sum += r
		st.count += 1
		st.timings[r] += 1
	}
	return nil
}

func (st *StatsCollector) Median() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if st.count == 0 {
		return 0
	}

	half := st.count / 2
	for i := 0; i < len(st.timings); i++ {
		half -= st.timings[i]
		if half < 0 {
			if st.count%2 == 1 {
				return float64(i)
			}
			for j := i - 1; j >= 0; j-- {
				if st.timings[j] != 0 {
					return float64(i+j) / 2
				}
			}
			return float64(i)
		}
	}
	return 0
}

func (st *StatsCollector) Average() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if st.count == 0 {
		return 0
	}
	return float64(st.sum) / float64(st.count)
}

func main() {

}
