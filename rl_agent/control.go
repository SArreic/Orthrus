package rl_agent

import (
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

// 修改 assign 逻辑：你需要手动去 request/request.go 里这样改

// 原始实现：
/*
func GetBucketNr(clID int32, clSN int32, senderId int32) int {
    return int(senderId % int32(config.Config.NumBuckets))
}
*/

// 改为支持 RL 控制映射
/*
func GetBucketNr(clID int32, clSN int32, senderId int32) int {
    idx := int(senderId % int32(len(request.Buckets)))
    if len(bucketMap) == len(request.Buckets) {
        return bucketMap[idx]
    }
    return idx
}
*/
