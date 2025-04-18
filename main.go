package main

import (
	"fmt"
	"time"

	"github.com/Hanzheng2021/orthrus/rl_agent"
	"github.com/Hanzheng2021/orthrus/request"
	"github.com/Hanzheng2021/orthrus/routing"
)

func ensureBuckets() {
	if len(request.Buckets) < 4 {
		for i := len(request.Buckets); i < 4; i++ {
			request.Buckets = append(request.Buckets, request.NewBucket(i))
		}
	}
}

func main() {
	ensureBuckets()

	fmt.Println("🚀 Starting RL Agent test client...")

	// 采集状态
	state := routing.CollectState(4) // 假设初始有4个实例
	fmt.Println("📡 Sending state:", state)

	// 与Python端通信
	action, err := rl_agent.QueryRLAction(state)
	if err != nil {
		fmt.Println("❌ Error communicating with RL agent:", err)
		return
	}

	// 输出动作
	fmt.Println("✅ Received action:", action.Type)

	// 应用动作
	switch action.Type {
	case 0:
		rl_agent.AddInstance()
	case 1:
		rl_agent.RemoveInstance()
	case 2:
		rl_agent.ReassignBuckets()
	}

	// 观察效果
	fmt.Println("📦 Current instance count:", len(rl_agent.GetAllBuckets()))
	time.Sleep(1 * time.Second)
}
