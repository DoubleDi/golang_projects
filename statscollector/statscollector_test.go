package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const requestTimeout = 19000

func TestCollector(t *testing.T) {
	testCases := []struct {
		values  []int64
		average float64
		median  float64
		err     error
	}{
		{
			values:  []int64{0, 0, 3, 19000, 19000},
			average: 7600.6,
			median:  3,
		},
		{},
		{
			values:  []int64{1, 1, 3, 3},
			average: 2,
			median:  2,
		},
		{
			values:  []int64{1, 1, 1, 1, 3, 2},
			average: 1.5,
			median:  1,
		},
		{
			values:  []int64{6, 6, 1, 1},
			average: 3.5,
			median:  3.5,
		},
		{
			values: []int64{requestTimeout + 100},
			err:    ErrRecordMoreThanTimeout,
		},
	}

	for i, tc := range testCases {
		collector := NewStatsCollector(requestTimeout)
		if tc.err != nil {
			require.Error(t, ErrRecordMoreThanTimeout, collector.Record(tc.values...), "%d", i)
		} else {
			require.NoError(t, collector.Record(tc.values...), "%d", i)
			assert.EqualValues(t, tc.average, collector.Average(), "%d", i)
			assert.EqualValues(t, tc.median, collector.Median(), "%d", i)
		}
	}
}

func BenchmarkMedian(b *testing.B) {
	collector := NewStatsCollector(requestTimeout)

	s1 := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s1)

	for i := 0; i < b.N; i++ {
		collector.Record(r.Int63n(requestTimeout))
		collector.Median()
	}
}

func BenchmarkAverage(b *testing.B) {
	collector := NewStatsCollector(requestTimeout)

	s1 := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s1)

	for i := 0; i < b.N; i++ {
		collector.Record(r.Int63n(requestTimeout))
		collector.Average()
	}
}
