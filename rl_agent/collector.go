package rl_agent

import (
	"sync"
	"time"
	"math/rand"

	"github.com/Hanzheng2021/orthrus/request"
)

type State struct {
	CPUUtilization       []float64 // 每个实例的CPU使用率（模拟）
	QueueLengths         []int     // 每个bucket当前排队交易数
	CrossInstanceRatio   float64   // 模拟跨实例交易比例
	AvgNetworkLatency    float64   // 模拟网络延迟（P95滑动均值）
	Timestamp            time.Time // 当前状态的时间戳
}

// StateCollector 负责定期收集系统状态
type StateCollector struct {
	mu      sync.Mutex
	state   State
	numInst int // 实例数（或bucket数）
}

// 创建新的状态收集器
func NewStateCollector(numInstances int) *StateCollector {
	return &StateCollector{
		numInst: numInstances,
	}
}

// Collect 从系统中采集最新状态，更新内部state字段
func (sc *StateCollector) Collect() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.state = State{
		CPUUtilization:     getMockCPU(sc.numInst),
		QueueLengths:       getQueueLengths(),
		CrossInstanceRatio: mockCrossInstanceRatio(),
		AvgNetworkLatency:  mockNetworkLatency(),
		Timestamp:          time.Now(),
	}
}

func CollectState(numInstances int) State {
	cpu := make([]float64, numInstances)
	for i := range cpu {
		cpu[i] = 30 + rand.Float64()*60
	}
	queues := make([]int, len(request.Buckets))
	for i, b := range request.Buckets {
		queues[i] = b.Len()
	}
	return State{
		CPUUtilization:     cpu,
		QueueLengths:       queues,
		CrossInstanceRatio: rand.Float64() * 0.3,
		AvgNetworkLatency:  20 + rand.Float64()*30,
	}
}

// 返回上一次采集到的状态
func (sc *StateCollector) GetState() State {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.state
}

// ========= 模拟指标获取函数，可后续替换为真实插桩接口 ==========

// 模拟 CPU 使用率（可改用 gopsutil 采集真实数据）
func getMockCPU(n int) []float64 {
	cpu := make([]float64, n)
	for i := range cpu {
		cpu[i] = 30 + rand.Float64()*50 // 模拟 30% ~ 80%
	}
	return cpu
}

// 获取所有 bucket 当前长度（真实数据）
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


// 模拟跨实例交易比例（真实实现需要你记录交易分配历史）
func mockCrossInstanceRatio() float64 {
	return rand.Float64() * 0.3 // 假设最多30%为跨实例
}

// 模拟网络延迟（真实实现建议在 messenger 模块插桩记录往返时间）
func mockNetworkLatency() float64 {
	return 30 + rand.Float64()*40 // 模拟 30~70ms 的 P95 延迟
}

// 你可以继续：
//     替换 getMockCPU() 为真实系统指标采集；
//     在交易分发路径中统计实际的 crossInstanceRatio；
//     在 messenger 层或请求收发链路上采样 network latency；
// 如果你需要，我可以：
//     帮你构建 RLAgent.Decide(state State) Action 的接口；
//     帮你实现 ApplyAction() 模块；
//     编写 RL 主控循环与训练框架。