package metrics

import (
    "sync"
    "time"
)

type Recorder struct {
    mu            sync.Mutex
    RequestCount  int
    LatencyHist   []time.Duration
}

func New() *Recorder {
    return &Recorder{
        LatencyHist: make([]time.Duration, 0),
    }
}

func (r *Recorder) RecordRequest(latency time.Duration) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.RequestCount++
    r.LatencyHist = append(r.LatencyHist, latency)
}

func (r *Recorder) GetAvgLatency() time.Duration {
    total := time.Duration(0)
    for _, l := range r.LatencyHist {
        total += l
    }
    return total / time.Duration(len(r.LatencyHist))
}

var (
    throughput        float64
    avgLatency        float64
    crossInstanceRate float64
    mu                sync.Mutex
)

func GetThroughput() float64 {
    mu.Lock()
    defer mu.Unlock()
    return throughput
}

func GetAvgLatency() float64 {
    mu.Lock()
    defer mu.Unlock()
    return avgLatency
}

func GetCrossInstanceRate() float64 {
    mu.Lock()
    defer mu.Unlock()
    return crossInstanceRate
}

func UpdateThroughput(t float64) {
    mu.Lock()
    defer mu.Unlock()
    throughput = t
}

var CrossInstanceCounter = &counterImpl{}

type counterImpl struct {
    mu    sync.Mutex
    count int64
}

func (c *counterImpl) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}