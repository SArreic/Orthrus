package rl_agent

import (
	"fmt"
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
	ReassignBucketsToActivePeers()
}

// 动作 1：移除一个 Orderer 实例（保留至少一个）
func RemoveOrdererInstance() {
	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("❌ Global MirManager not initialized.")
		return
	}

	var lastPeerID int32 = -1
	mm.ForEachPeer(func(id int32) {
		if id >= 1000 && id > lastPeerID {
			lastPeerID = id
		}
	})

	if lastPeerID == -1 {
		fmt.Println("⚠️ No removable dynamic Orderer instance.")
		return
	}

	fmt.Println("🗑 Removing dynamic Orderer:", lastPeerID)

	messenger.DisconnectPeer(lastPeerID)
	membership.UnregisterNode(lastPeerID)
	mm.UnregisterOrderer(lastPeerID)

	if len(request.Buckets) > 0 {
		request.Buckets = request.Buckets[:len(request.Buckets)-1]
	}
	ReassignBucketsToActivePeers()
}

// 动作 2：重新分配桶映射
var bucketMap []int

func ReassignBucketsToActivePeers() {
	activePeers := []int32{}
	manager.GetGlobalMirManager().ForEachPeer(func(id int32) {
		activePeers = append(activePeers, id)
	})
	numPeers := len(activePeers)
	numBuckets := len(request.Buckets)

	newMap := make([]int, numBuckets)
	for i := 0; i < numBuckets; i++ {
		newMap[i] = i % numPeers
	}
	routing.SetBucketMap(newMap)
	bucketMap = newMap
	fmt.Println("✅ Reassigned bucket mappings to active peers:", newMap)
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
		RemoveOrdererInstance()
	case 2:
		fmt.Println("⏸️ No-op action.")
	case 3:
		ReassignBucketsToActivePeers()
	default:
		fmt.Println("⚠️ Unknown action:", act.Type)
	}
}

