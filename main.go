package main

import (
    "fmt"
    "github.com/Hanzheng2021/orthrus/manager"
    "time"
)

func main() {
    rl := &manager.RLClient{
        Endpoint: "http://127.0.0.1:5000/decide",
        Timeout:  3 * time.Second,
    }

    testState := manager.RLState{
        Epoch:    1,
        Orderers: []int32{0, 1, 2},
        BucketStats: []manager.BucketStat{
            {ID: 0, TxCount: 100, CrossRatio: 0.2},
            {ID: 1, TxCount: 50, CrossRatio: 0.1},
            {ID: 2, TxCount: 75, CrossRatio: 0.3},
            {ID: 3, TxCount: 0, CrossRatio: 0.0},
        },
        OrdererStats: []manager.OrdererStat{
            {ID: 0, CPU: 0.5, Latency: 100},
            {ID: 1, CPU: 0.4, Latency: 110},
            {ID: 2, CPU: 0.3, Latency: 90},
        },
    }

    assignment, err := rl.DecideAssignment(testState)
    if err != nil {
        fmt.Println("Error calling RL agent:", err)
    } else {
        fmt.Println("Assignment from RL agent:", assignment)
    }
}
