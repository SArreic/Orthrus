// routing/state.go
package routing

import "time"

type State struct {
	CPUUtilization     []float64 // 每个实例的CPU使用率
	QueueLengths       []int     // 每个bucket当前排队交易数
	CrossInstanceRatio float64   // 跨实例交易比例
	AvgNetworkLatency  float64   // 网络延迟（P95滑动均值）
	Timestamp          time.Time // 当前状态的时间戳
}
