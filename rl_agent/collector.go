package rl_agent

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/Hanzheng2021/orthrus/routing"
	"github.com/Hanzheng2021/orthrus/request"
)

func init() {
	routing.SetLatencyRecorder(latencyCollector{})
	routing.SetCrossInstanceRecorder(crossRecorder{})
	routing.SetStateCollector(stateCollector{})
}

// 收集最新状态（暴露给主调度器调用）

type stateCollector struct{}

func (stateCollector) CollectState(numInstances int) routing.State {
	return routing.State{
		CPUUtilization:     getCPUUtilization(numInstances),
		QueueLengths:       getQueueLengths(),
		CrossInstanceRatio: getCrossInstanceRatio(),
		AvgNetworkLatency:  getAvgLatency(),
		Timestamp:          time.Now(),
	}
}

// ===================== CPU 使用率采集 =====================

func getCPUUtilization(n int) []float64 {
	percentages, err := cpu.Percent(0, true) // 获取每个逻辑核的瞬时占用率
	if err != nil || len(percentages) == 0 {
		return fallbackCPU(n)
	}
	if len(percentages) >= n {
		return percentages[:n]
	}
	return append(percentages, fallbackCPU(n-len(percentages))...)
}

func fallbackCPU(n int) []float64 {
	cpu := make([]float64, n)
	for i := 0; i < n; i++ {
		cpu[i] = 30 + float64(i*7%20) // 假定值，避免报错
	}
	return cpu
}

// ===================== Bucket 队列长度 =====================

func getQueueLengths() []int {
	buckets := request.Buckets
	n := len(buckets)
	if n == 0 {
		return []int{0, 0, 0, 0}
	}
	lengths := make([]int, n)
	for i, b := range buckets {
		lengths[i] = b.Len()
	}
	return lengths
}

// ===================== 网络延迟统计 =====================

var latencyWindow []float64
var latencyMu sync.Mutex
const latencyWindowSize = 50

type latencyCollector struct{}

func (latencyCollector) RecordLatency(ms int64) {
	latencyMu.Lock()
	defer latencyMu.Unlock()
	if len(latencyWindow) >= latencyWindowSize {
		latencyWindow = latencyWindow[1:]
	}
	latencyWindow = append(latencyWindow, float64(ms))
}

func getAvgLatency() float64 {
	latencyMu.Lock()
	defer latencyMu.Unlock()
	if len(latencyWindow) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range latencyWindow {
		sum += v
	}
	return sum / float64(len(latencyWindow))
}

// ===================== 跨实例交易比例统计 =====================

var totalTx int
var crossTx int
var txMu sync.Mutex

type crossRecorder struct{}

func (crossRecorder) RecordCrossInstance(cross bool) {
	txMu.Lock()
	defer txMu.Unlock()
	totalTx++
	if cross {
		crossTx++
	}
	if totalTx > 1000 {
		totalTx = totalTx / 2
		crossTx = crossTx / 2
	}
}

func getCrossInstanceRatio() float64 {
	txMu.Lock()
	defer txMu.Unlock()
	if totalTx == 0 {
		return 0
	}
	return float64(crossTx) / float64(totalTx)
}