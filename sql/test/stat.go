package test

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type stat struct {
	latency []int64
	count   int64
	slow    int64
	maxSlow int64
}

func (s *stat) reset(msSlowTime int64) {
	s.latency = make([]int64, msSlowTime)
	s.count = 0
	s.slow = 0
	s.maxSlow = 0
}

func (s *stat) add(other *stat) {
	n := len(other.latency)
	if n != len(s.latency) {
		panic("invalid stat latency")
	}

	for ms := 0; ms < n; ms++ {
		s.latency[ms] += other.latency[ms]
	}

	s.count += other.count
	s.slow += other.slow
	if s.maxSlow < other.maxSlow {
		s.maxSlow = other.maxSlow
	}

}

type StatManager struct {
	concurrency int64
	methods     []string
	period      int64
	ratios      []float32
	slowTime    int64
	doing       [][]stat
	done        [][]*stat
	lock        []sync.Mutex
	endTime     atomic.Int64
	idSeed      atomic.Int64
}

func NewStatManager(concurrency, period int64, methods []string, ratios []float32, slowTime int64) *StatManager {
	ret := &StatManager{
		concurrency: concurrency,
		methods:     methods,
		period:      period,
		ratios:      ratios,
		slowTime:    slowTime,
		doing:       make([][]stat, concurrency),
		done:        make([][]*stat, concurrency),
		lock:        make([]sync.Mutex, concurrency),
	}

	nMethod := len(methods)
	for i := int64(0); i < concurrency; i++ {
		ret.doing[i] = make([]stat, nMethod)
		ret.done[i] = make([]*stat, nMethod)

		for j := 0; j < nMethod; j++ {
			ret.doing[i][j].reset(slowTime)
		}
	}

	return ret
}

func (sm *StatManager) NewIndex() int64 {
	return sm.idSeed.Add(1) - 1
}

func (sm *StatManager) Add(id int64, method int64, startTime int64, endTime int64) {
	if id < 0 || id >= sm.concurrency {
		panic("invalid id")
	}

	doingStat := &sm.doing[id][method]
	doingStat.count++
	ms := (endTime - startTime) / int64(time.Millisecond)
	if ms < 0 {
		ms = 0
	}

	if ms >= sm.slowTime {
		doingStat.slow++
		if ms > doingStat.maxSlow {
			doingStat.maxSlow = ms
		}

	} else {
		doingStat.latency[ms]++
	}

	if endTime > sm.endTime.Load() {
		s := &stat{latency: doingStat.latency, count: doingStat.count, slow: doingStat.slow, maxSlow: doingStat.maxSlow}
		doingStat.reset(sm.slowTime)

		sm.lock[id].Lock()
		doneStat := sm.done[id][method]
		if doneStat == nil {
			sm.done[id][method] = s

		} else {
			doneStat.add(s)
		}
		sm.lock[id].Unlock()
	}
}

func (sm *StatManager) Run() {
	currentTime := time.Now().UTC().UnixNano()
	sm.endTime.Store(currentTime + (sm.period-3)*int64(time.Second))

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(sm.period))
		for range ticker.C {
			fmt.Printf("============= %s =============\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
			nMethod := len(sm.methods)
			stats := make([][]*stat, nMethod)
			for i := 0; i < nMethod; i++ {
				stats[i] = make([]*stat, sm.concurrency)
			}

			done := sm.done
			lock := sm.lock
			for i := int64(0); i < sm.concurrency; i++ {
				lock[i].Lock()
				for m := 0; m < nMethod; m++ {
					stats[m][i] = done[i][m]
					done[i][m] = nil
				}
				lock[i].Unlock()
			}

			for m := 0; m < nMethod; m++ {
				sum := new(stat)
				sum.reset(sm.slowTime)
				for i := int64(0); i < sm.concurrency; i++ {
					stat := stats[m][i]
					if stat == nil {
						continue
					}

					sum.add(stat)
				}

				if sum.count == 0 {
					continue
				}

				nRatio := int64(len(sm.ratios))
				cntRatio := make([]int64, nRatio)
				for i := int64(0); i < nRatio; i++ {
					cntRatio[i] = int64(float32(sum.count) * sm.ratios[i])
				}

				rc := int64(0)
				ri := int64(0)
				var builder strings.Builder
				builder.WriteString(fmt.Sprintf(
					"%s: num=%d; slow=%d; max_slow=%dms; ratio=",
					sm.methods[m], sum.count, sum.slow, sum.maxSlow))
				prev := "{"
				for ms, cnt := range sum.latency {
					if cnt == 0 {
						continue
					}

					rc += cnt
					for i := ri; i < nRatio; i++ {
						if rc < cntRatio[i] {
							continue
						}

						builder.WriteString(prev)
						builder.WriteString(fmt.Sprintf("%.2f: %dms", sm.ratios[i], ms))
						prev = "; "
						ri = i + 1
						break
					}
				}
				builder.WriteString("}\n")
				fmt.Printf(builder.String())
			}

			sm.endTime.Add(sm.period * int64(time.Second))
			fmt.Printf("---------------------------------------------------\n")
		}
	}()
}
