package rl_agent

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Hanzheng2021/orthrus/manager"
	"github.com/Hanzheng2021/orthrus/membership"
	"github.com/Hanzheng2021/orthrus/messenger"
	"github.com/Hanzheng2021/orthrus/request"
	"github.com/Hanzheng2021/orthrus/routing"
	pb "github.com/Hanzheng2021/orthrus/protobufs"
	logger "github.com/rs/zerolog/log"
)

var (
	peerIDCounter int32 = -1 // 注意：在系统启动后动态初始化！
	once          sync.Once
	epoch         int32 = 0
)

// 初始化peerIDCounter
func InitPeerIDCounterFromSystem() {
	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("❌ MirManager not initialized, cannot initialize PeerIDCounter.")
		return
	}
	maxID := int32(-1)
	mm.ForEachPeer(func(id int32) {
		if id > maxID {
			maxID = id
		}
	})
	peerIDCounter = maxID
	fmt.Println("🔧 Initialized peerIDCounter to", peerIDCounter)
}

// 生成新的PeerID
func generateNextPeerID() int32 {
	peerIDCounter++
	return peerIDCounter
}

// 动作0：增加一个Orderer实例
func AddInstance() {
	fmt.Println("Entered AddInstance()")
	newID := generateNextPeerID()

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

// 动作1：移除一个Orderer实例（保留至少一个）
func RemoveOrdererInstance() {
	fmt.Println("Entered RemoveOrdererInstance()")

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

	// 移除Orderer后，重新分配桶
	ReassignBucketsToActivePeers()
	AnnounceBucketsToClients()
}

// 动作2：重新分配桶
func ReassignBucketsToActivePeers() {
	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("⚠️ ReassignBucketsToActivePeers: MirManager not ready, skip this tick.")
		return
	}

	activePeers := []int32{}
	mm.ForEachPeer(func(id int32) {
		activePeers = append(activePeers, id)
	})

	numPeers := len(activePeers)
	numBuckets := len(request.Buckets)
	fmt.Println("Number of active peers: %d", numPeers)
	fmt.Println("Number of Buckets: %d", numBuckets)

	if numPeers == 0 || numBuckets == 0 {
		logger.Error().Msg("Reassigning Buckets but number of active peers or buckets is 0!")
		return
	}

	newMap := make([]int, numBuckets)
	for i := 0; i < numBuckets; i++ {
		newMap[i] = int(activePeers[i%numPeers])
		fmt.Println("")
	}

	routing.SetBucketMap(newMap)
	fmt.Println("✅ Reassigned bucket mappings to active peers:", newMap)
}

// 启动 RL 控制循环
func StartRLControlLoop() {
	once.Do(func() {
		InitPeerIDCounterFromSystem() // 加上启动时同步
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

// 根据动作类型分发调用
func applyAction(act Action) {
	fmt.Println("🧹 applyAction called, action type =", act.Type)
	logger.Info().Int("ActionType", int(act.Type)).Msg("🧹 applyAction called.")

	switch act.Type {
	case 0:
		fmt.Println("➕ Trying to add instance...")
		AddInstance()
	case 1:
		if manager.GetGlobalMirManager().CountDynamicOrderers() > 0 {
			fmt.Println("➖ Trying to remove instance...")
			RemoveOrdererInstance()
		} else {
			fmt.Println("⚠️ No dynamic orderer to remove, fallback to no-op.")
		}
	case 2:
		fmt.Println("🔄 Reassigning buckets...")
		ReassignBucketsToActivePeers()
	case 3:
		fmt.Println("⏸️ No-op action received.")
	default:
		fmt.Println("⚠️ Unknown action type:", act.Type)
	}
}

// 广播当前Buckets给客户端
func AnnounceBucketsToClients() {
	fmt.Println("📣 Enter AnnounceBucketsToClients()")
	logger.Info().Msg("📣 Enter AnnounceBucketsToClients()")
	os.Stdout.Sync()

	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("❌ MirManager not initialized.")
		logger.Error().Msg("❌ MirManager not initialized.")
		os.Stdout.Sync()
		return
	}

	numBuckets := len(request.Buckets)
	fmt.Println("🗂️ request.Buckets count:", numBuckets)
	if numBuckets == 0 {
		fmt.Println("❌ No buckets available to assign.")
		return
	}

	// 通过Manager重新分配
	bucketsMap := mm.AssignBuckets(numBuckets)

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

	fmt.Println("📣 Announcing BucketAssignment to clients: %+v", assignment)
	messenger.AnnounceBucketAssignment(assignment)
	os.Stdout.Sync()
}