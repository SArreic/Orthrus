package rl_agent

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Hanzheng2021/orthrus/manager"
	"github.com/Hanzheng2021/orthrus/membership"
	"github.com/Hanzheng2021/orthrus/messenger"
	"github.com/Hanzheng2021/orthrus/request"
	"github.com/Hanzheng2021/orthrus/routing"
	pb "github.com/Hanzheng2021/orthrus/protobufs"
)

var peerIDCounter int32 = 1000

func generateNextPeerID() int32 {
	peerIDCounter++
	return peerIDCounter
}

func GetBucketMap() []int {
	return bucketMap
}

// 动作 0：增加一个 Orderer 实例
func AddInstance() {
	newID := generateNextPeerID()

	newBucket := request.NewBucket(int(newID))
	request.Buckets = append(request.Buckets, newBucket)

	identity := &pb.NodeIdentity{
		NodeId:      newID,
		PublicAddr:  "127.0.0.1",
		PrivateAddr: "127.0.0.1",
		Port:        newID + 6000,
	}

	membership.RegisterNewNode(identity)
	manager.GetGlobalMirManager().RegisterNewOrderer(newID)
	messenger.ConnectToPeer(identity)

	fmt.Println("✅ Added new Orderer instance:", newID)
}

// 动作 1：移除一个 Orderer 实例（保留至少一个）
func RemoveInstance() {
	n := len(request.Buckets)
	if n <= 1 {
		fmt.Println("⚠️ Cannot remove instance, only one left.")
		return
	}
	request.Buckets = request.Buckets[:n-1]
	fmt.Println("✅ Removed instance/bucket:", n-1)
}

// 动作 2：重新分配桶映射
var bucketMap []int

func ReassignBuckets() {
	numBuckets := len(request.Buckets)
	bmap := make([]int, numBuckets)
	for i := 0; i < numBuckets; i++ {
		bmap[i] = rand.Intn(numBuckets)
	}
	routing.SetBucketMap(bmap)
	bucketMap = bmap
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
				state := routing.CollectState(len(request.Buckets))
				action, err := QueryRLAction(state)
				if err != nil {
					fmt.Println("❌ RL agent query failed:", err)
					time.Sleep(2 * time.Second)
					continue
				}
				applyAction(action)
				time.Sleep(5 * time.Second)
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
