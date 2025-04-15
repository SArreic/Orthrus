// Copyright 2022 IBM Corp. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package request

import (
	// "crypto/sha256"

	"log"
	"strconv"

    "sync"
	"time"
    "github.com/Hanzheng2021/orthrus/rl_agent"

	"github.com/golang/protobuf/proto"
	"github.com/Hanzheng2021/orthrus/config"
	"github.com/Hanzheng2021/orthrus/tracing"

	"github.com/Hanzheng2021/orthrus/messenger"
    "github.com/Hanzheng2021/orthrus/metrics"

	// "github.com/Hanzheng2021/orthrus/crypto"
	pb "github.com/Hanzheng2021/orthrus/protobufs"
)

var rlAgent = rl_agent.NewDQNAgent(
    len(Buckets) + 2,  // 位置参数
    len(Buckets),
)

// TODO: It's inefficient to hash a request every time it is needed to get the request ID

// This function is used by the messenger as the handler function for requests (the main file performs the assignment).
// Simply adds the received request to the corresponding request buffer.
// TODO: If too many threads (64 or more in the current deployment with 32-core machines) invoke Add(),
//
//	the buffer locks get extremely contended.
//	Have only a fixed (configurable) number of threads invoking Add().
//	Spawn those worker threads in the Init() function and make HandleRequest (this function) only write
//	the request to a channel (do we need a big channel buffer for this?) that a worker reads.
//	It would make sense to send requests from the same client to the same worker,
//	Since the Buffer lock to be acquired by the worker is determined by the clientID.
//	The lock being acquired by the same thread is crucial for avoiding contention.
//	If this is not enough, try having the worker threads add requests to buffers in batches.
//	(Although this might be very tricky if we want to avoid verifying signatures while holding the buffer lock,
//	and at the same time avoid verifying the signature again, in case the request is already present.)
func HandleRequest(req *pb.ClientRequest) {

	tracing.Trace2.EventForClientInPeer(tracing.REQ_RECEIVE, int64(req.RequestId.ClientSn), req.RequestId.ClientId)

	// 新增RL逻辑
    state := collectState()
    action := rlAgent.SelectAction(state)
    targetBucket := action.TargetBucket
    Buckets[targetBucket].Add(req)

    // 定期训练（需导入time包）
    if time.Now().Unix()%10 == 0 {
        go rlAgent.Train()
    }

	if config.Config.RequestHandlerThreads > 0 {
		// Write request to the corresponding input channel for further processing by a request handler thread.
		// There is a fixed number of request handler threads (should be at most as many as there are physical cores)
		// to avoid cache contention on the request Buffers. To avoid this contention, it is also crucial that requests from
		// the same client (there is a separate Buffer per client) are handled by the same request handler thread.
		requestInputChannels[int(req.RequestId.ClientId)%config.Config.RequestHandlerThreads] <- req
	} else {
		AddReqMsg(req)
	}

    // 添加请求时记录发送者桶信息
    req.SenderBucket = int32(GetBucketByHashing(req).id) // 需要protobuf添加该字段
    
    Buckets[targetBucket].Add(req) // 使用新的Add方法
    trackCrossInstance(req, targetBucket)
}

func GetBucketByHashing(req *pb.ClientRequest) *Bucket {

	newTx := &pb.Transaction{}
	err := proto.Unmarshal(req.Payload, newTx)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
	}

	SID, err := strconv.Atoi(newTx.SenderHash)
	if err != nil {
		log.Fatal("GetBucketByHashing error: ", err)
	}

	// fmt.Printf("bucket index i is %d\n", i)
	b := Buckets[SID%config.Config.NumBuckets]

	return b
}

func GetBucketByRL(state rl_agent.SystemState, req *pb.ClientRequest) *Bucket {
    action := rlAgent.SelectAction(state)
    return Buckets[action.TargetBucket]
}

// 修改requesthandler.go中的getBucketLoads()
func getBucketLoads() []float64 {
    loads := make([]float64, len(Buckets))
    for i, b := range Buckets {
        b.Lock()
        loads[i] = float64(b.numRequests) // 直接读取计数
        b.Unlock()
    }
    return loads
}

func collectState() rl_agent.SystemState {
    return rl_agent.SystemState{
        BucketLoads:       getBucketLoads(),
        Throughput:        metrics.GetThroughput(),
        AvgLatency:        metrics.GetAvgLatency(),
        NetworkLatency:    messenger.GetAverageLatency(),
        CrossInstanceRate: metrics.GetCrossInstanceRate(),
    }
}

func trackCrossInstance(req *pb.ClientRequest, targetBucket int) {
    senderBucket := req.GetSenderBucket() // 使用Get方法安全访问
    if senderBucket != int32(targetBucket) {
        if metrics.CrossInstanceCounter != nil {
            metrics.CrossInstanceCounter.Inc()
        }
    }
}

var (
    bucketsMu    sync.Mutex
    globalBuckets []*Bucket
    globalRLAgent = rl_agent.NewDQNAgent(0, 0)
)

func InitRequestHandler() {
    bucketsMu.Lock()
    defer bucketsMu.Unlock()
    
    globalBuckets = make([]*Bucket, config.Config.NumBuckets)
    for i := range globalBuckets {
        globalBuckets[i] = NewBucket(i)
    }
    globalRLAgent = rl_agent.NewDQNAgent(len(globalBuckets)+2, len(globalBuckets))
}
