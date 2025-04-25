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

var peerIDCounter int32 = -1

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

	numBuckets := len(request.Buckets) + 1
	request.InitBuckets(numBuckets)

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
	AnnounceBucketsToClients()
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
	AnnounceBucketsToClients()
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

	fmt.Printf("🔎 Buckets count: %d\n", len(request.Buckets))
	fmt.Printf("📊 New bucketMap: %v\n", newMap)

	for i := 0; i < numBuckets; i++ {
		newMap[i] = i % numPeers
	}
	routing.SetBucketMap(newMap)
	bucketMap = newMap
	fmt.Println("✅ Reassigned bucket mappings to active peers:", newMap)
	
	// 🧹 清理过期桶或映射（防止误用旧桶）
	if len(request.Buckets) != len(newMap) {
		fmt.Printf("⚠️ BucketMap/Bucket length mismatch! Buckets=%d Map=%d\n", len(request.Buckets), len(newMap))
	}

	AnnounceBucketsToClients()
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

var epoch int32 = 0

func AnnounceBucketsToClients() {
	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("❌ MirManager not initialized.")
		return
	}

	numBuckets := len(request.Buckets)
	if numBuckets == 0 {
		fmt.Println("❌ No buckets available to assign.")
		return
	}

	bucketsMap := mm.AssignBuckets(numBuckets)
	fmt.Printf("📣 Buckets to announce:\n")
	for peer, list := range bucketsMap {
		fmt.Printf("  Peer %d => Buckets %v\n", peer, list)
	}

	pbBuckets := make(map[int32]*pb.ListOfInt32)
	for peerID, buckets := range bucketsMap {
		pbBuckets[peerID] = &pb.ListOfInt32{Vals: make([]int32, len(buckets))}
		for i, b := range buckets {
			pbBuckets[peerID].Vals[i] = int32(b)
		}
	}

	epoch++
	assignment := &pb.BucketAssignment{
		Epoch:   epoch,
		Buckets: pbBuckets,
	}

	fmt.Printf("📣 Announcing BucketAssignment to clients: %+v\n", assignment)
	messenger.AnnounceBucketAssignment(assignment)
}