package rl_agent

import (
	"sync"
	"time"
	"fmt"
	"math/rand"

	"github.com/Hanzheng2021/orthrus/request"
	"github.com/Hanzheng2021/orthrus/routing"
)

// var bucketMap []int

func GetBucketMap() []int {
	return bucketMap
}

// 动作 0：增加一个 bucket/实例
func AddInstance() {
	newID := len(request.Buckets)
	newBucket := request.NewBucket(newID)
	request.Buckets = append(request.Buckets, newBucket)
	fmt.Println("✅ Added new instance/bucket:", newID)
}

// 动作 1：移除最后一个 bucket（前提是还有多个）
func RemoveInstance() {
	n := len(request.Buckets)
	if n <= 1 {
		fmt.Println("⚠️ Cannot remove instance, only one left.")
		return
	}
	// 可选：处理 bucket[n-1] 中的遗留请求
	request.Buckets = request.Buckets[:n-1]
	fmt.Println("✅ Removed instance/bucket:", n-1)
}

// 动作 2：重新分配桶到实例的映射（这里是简单地打乱）
var bucketMap []int

func ReassignBuckets() {
	numBuckets := len(request.Buckets)
	bmap := make([]int, numBuckets)
	for i := 0; i < numBuckets; i++ {
		bmap[i] = rand.Intn(numBuckets)
	}
	routing.SetBucketMap(bmap)
	fmt.Println("✅ Reassigned bucket mappings:", bmap)
}

func GetAllBuckets() []*request.Bucket {
	return request.Buckets
}

var once sync.Once

func StartRLControlLoop() {
	once.Do(func() {
		go func() {
			for {
				// 获取当前状态（你通过 routing.SetStateCollector 注册的逻辑）
				state := routing.CollectState(len(request.Buckets)) // 你可能需要实现这个函数或把原有收集逻辑导出

				// 获取动作
				action, err := QueryRLAction(state)
				if err != nil {
					fmt.Println("❌ RL agent query failed:", err)
					continue
				}

				// 应用动作
				applyAction(action)

				time.Sleep(5 * time.Second) // 控制策略应用频率
			}
		}()
	})
}

func applyAction(act Action) {
	switch act.Type {
	case 0:
		AddInstance()
	case 1:
		RemoveInstance()
	case 2:
		fmt.Println("⏸️ No-op action.")
	case 3:
		ReassignBuckets()
	default:
		fmt.Println("⚠️ Unknown action:", act.Type)
	}
}
