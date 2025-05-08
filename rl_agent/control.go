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
	peerIDCounter int32 = -1 // 初始化后同步
	once          sync.Once
	epoch         int32 = 0
)

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

	// 注册新节点
	membership.RegisterNewNode(identity)
	manager.GetGlobalMirManager().RegisterNewOrderer(newID)
	messenger.ConnectToPeer(identity)

	// 创建新桶并更新routing.BucketMap
	newBucket := request.NewBucket(int(newID))  // 假设NewBucket为创建新桶的函数
	request.Buckets = append(request.Buckets, newBucket)
	newBucketID := len(request.Buckets) - 1  // 新桶的ID

	// 更新routing.BucketMap，使得新桶优先接收请求
	currentBucketMap := routing.GetBucketMap()
	currentBucketMap = append(currentBucketMap, newBucketID)  // 将新桶加入分配映射

	// 设置新的桶分配映射
	routing.SetBucketMap(currentBucketMap)

	// 需要追踪request被添加到buckets中的具体逻辑
	// 然后设置tracker追踪次数，在新加入的桶负载
	// 达到一定值（比如均值）的时候再切断
	// 在那之前，新加入的request会全部加入到新桶

	// // 初始化一个变量，来追踪新桶的负载
	// newBucketLoad := 0
	// maxLoad := int(float64(len(request.Buckets)) / float64(len(manager.GetGlobalMirManager().Peers())) * float64(config.Config.MaxRequestsPerBucket))

	// // 循环检查桶负载，直到新桶达到最大负载
	// for newBucketLoad < maxLoad {
	// 	// 给新桶分配请求
	// 	// 比如，我们可以创建新的请求并直接将它们添加到新桶
	// 	newRequest := createNewRequest(newID)  // 创建新的请求
	// 	newBucket.AddRequest(newRequest)

	// 	// 更新负载
	// 	newBucketLoad++
	// 	if newBucketLoad >= maxLoad {
	// 		break  // 一旦达到负载上限，就停止分配给新桶
	// 	}
	// }

	// 当新桶填充满后，恢复常规的请求分配
	ReassignBucketsToActivePeers()
	AnnounceBucketsToClients()

	fmt.Println("✅ Added new Orderer instance and bucket filled.")
}


// 动作1：移除Orderer实例
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

	// 删除编号最大的实例
	messenger.DisconnectPeer(lastPeerID)
	membership.UnregisterNode(lastPeerID)
	mm.UnregisterOrderer(lastPeerID)

	// 更新peerIDCounter，确保新的实例编号正确
	peerIDCounter = lastPeerID - 1 // 更新为当前最大的peerID

	// 请求重新均衡分配到其他实例
	ReassignRequestsAfterRemoval(lastPeerID)
	AnnounceBucketsToClients()
}

// 动作2：重新分配桶（优先分配给负载较低的桶）
func ReassignBucketsToActivePeers() {
	mm := manager.GetGlobalMirManager()
	if mm == nil {
		fmt.Println("⚠️ MirManager not ready.")
		return
	}

	activePeers := []int32{}
	mm.ForEachPeer(func(id int32) {
		activePeers = append(activePeers, id)
	})

	numPeers := len(activePeers)
	numBuckets := len(request.Buckets)

	if numPeers == 0 || numBuckets == 0 {
		logger.Error().Msg("Reassigning Buckets but number of active peers or buckets is 0!")
		return
	}

	// 优先选择负载较低的桶
	loads := make(map[int]int) // 记录每个桶的当前负载
	for i := 0; i < numBuckets; i++ {
		bucketID := request.Buckets[i].GetId()
		loads[bucketID] = request.Buckets[i].Len()
	}

	// 根据负载情况重新分配桶
	newMap := make([]int, numBuckets)
	for i := 0; i < numBuckets; i++ {
		// 根据负载优先选择桶
		lowestLoadBucket := getLowestLoadBucket(loads)
		newMap[i] = lowestLoadBucket
		loads[lowestLoadBucket]++
	}

	routing.SetBucketMap(newMap)
	fmt.Println("✅ Reassigned bucket mappings to active peers with load balancing.")
}

func getLowestLoadBucket(loads map[int]int) int {
	// 选择负载最小的桶
	lowestLoad := int(^uint(0) >> 1) // 设置为最大整数
	var bucketID int
	for id, load := range loads {
		if load < lowestLoad {
			lowestLoad = load
			bucketID = id
		}
	}
	return bucketID
}

// 新增实例时请求分配到新实例
func ReassignRequestsToNewInstance(newID int32) {
	numBuckets := len(request.Buckets)
	for i := 0; i < numBuckets; i++ {
		request.Buckets[i].AddRequest(&request.Request{
			Msg: &pb.ClientRequest{
				RequestId: &pb.RequestID{
					ClientId: newID,
					ClientSn: int32(i),
				},
			},
		})
	}
}

func ReassignRequestsAfterRemoval(removedID int32) {
	activeBuckets := []int{}
	for _, b := range request.Buckets {
		if int32(b.GetId()) == removedID {
			activeBuckets = append(activeBuckets, b.GetId())
		}
	}
	// TODO: 这里可以使用循环方式确保负载均衡
}

func StartRLControlLoop() {
	once.Do(func() {
		InitPeerIDCounterFromSystem()

		go func() {
			for {
				mm := manager.GetGlobalMirManager()
				if mm == nil {
					fmt.Println("❌ MirManager not initialized. Waiting for initialization...")
					time.Sleep(2 * time.Second)
					continue
				}

				currentEpoch := mm.GetEpoch()

				if currentEpoch == 0 {
					fmt.Println("⏳ Waiting for epoch 0 to finish...")
					time.Sleep(2 * time.Second)
					continue
				}

				if currentEpoch != epoch {
					epoch = currentEpoch
					state := routing.CollectState(len(request.Buckets))
					action, err := QueryRLAction(state)
					if err != nil {
						fmt.Println("❌ RL agent query failed:", err)
						time.Sleep(2 * time.Second)
						continue
					}
					applyAction(action)
					logger.Info().Int32("epoch", epoch).Msg("Applied RL action for the new epoch")
				}

				time.Sleep(1 * time.Second)
			}
		}()
	})
}


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