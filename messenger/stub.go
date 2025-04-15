package messenger

import "time"

var (
    defaultLatency = 50 * time.Millisecond
)

func GetAverageLatency() float64 {
    return float64(defaultLatency.Milliseconds()) // 返回默认值
}