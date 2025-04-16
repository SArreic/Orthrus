package rl_agent

// type State struct {
// 	CPUUtilization     []float64 `json:"CPUUtilization"`
// 	QueueLengths       []int     `json:"QueueLengths"`
// 	CrossInstanceRatio float64   `json:"CrossInstanceRatio"`
// 	AvgNetworkLatency  float64   `json:"AvgNetworkLatency"`
// }

type Action struct {
	Type int `json:"Type"` // 0: add instance, 1: remove, 2: rebalance
}
