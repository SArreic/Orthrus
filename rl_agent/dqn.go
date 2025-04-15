// rl_agent/dqn.go
package rl_agent

import (
	"math/rand"
	"sync"
	"log"
)

type DQNAgent struct {
	policyNet     *PolicyNetwork
	targetNet     *PolicyNetwork
	replayBuffer  *ReplayBuffer
	epsilon       float64
	epsilonMin    float64
	epsilonDecay  float64
	mu            sync.Mutex
}

func NewDQNAgent(stateDim, actionDim int) *DQNAgent {
	return &DQNAgent{
		policyNet:    NewPolicyNetwork(stateDim, actionDim),
		targetNet:    NewPolicyNetwork(stateDim, actionDim),
		replayBuffer: NewReplayBuffer(10000),
		epsilon:      1.0,
		epsilonMin:   0.01,
		epsilonDecay: 0.995,
	}
}

func (a *DQNAgent) SelectAction(state SystemState) Action {
    a.mu.Lock()
    defer a.mu.Unlock()

    if rand.Float64() < a.epsilon {
        return Action{
            TargetBucket:    rand.Intn(len(state.BucketLoads)),
            ResizeInstances: rand.Intn(3) - 1,
        }
    }
    
    action := a.policyNet.Predict(state)
    
    log.Printf("RL Action - Epsilon: %.2f, Bucket: %d, Resize: %d",
        a.epsilon, action.TargetBucket, action.ResizeInstances)
    
    return action
}

func (a *DQNAgent) DecayEpsilon() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.epsilon > a.epsilonMin {
		a.epsilon *= a.epsilonDecay
	}
}

func (a *DQNAgent) Train() {
	if len(a.replayBuffer.buffer) < 1000 {
		return
	}
	
	batch := a.replayBuffer.Sample(32)
	// 这里实现DQN训练逻辑（需自行实现或调用深度学习库）
	// 伪代码示例：
	// loss := a.policyNet.Update(batch)
	_ = a.policyNet.Update(batch) // 用下划线接收返回值
	a.targetNet.SyncWeights(a.policyNet)
	
	a.DecayEpsilon()
}