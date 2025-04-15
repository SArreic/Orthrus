package rl_agent

import (
    "math"
)

func NormalizeState(raw SystemState) []float64 {
    features := make([]float64, 0)
    
    // 1. Bucket负载归一化
    if len(raw.BucketLoads) > 0 {
        maxLoad := getMax(raw.BucketLoads)
        for _, load := range raw.BucketLoads {
            features = append(features, load/maxLoad)
        }
    }

    // 2. 基础指标
    features = append(features,
        raw.Throughput/1000,       // 假设吞吐量最大为1000 TPS
        raw.AvgLatency/1000,       // 假设延迟最大为1000ms
    )

    // 3. 网络指标（可选）
    if raw.NetworkLatency > 0 {
        features = append(features, math.Tanh(raw.NetworkLatency/100))
    }
    
    // 4. 跨实例交易率
    features = append(features, raw.CrossInstanceRate)

    return features
}

// 修改为处理float64数组
func getMax(arr []float64) float64 {
    if len(arr) == 0 {
        return 1.0
    }
    
    max := arr[0]
    for _, v := range arr[1:] {
        if v > max {
            max = v
        }
    }
    return math.Max(max, 1.0)
}