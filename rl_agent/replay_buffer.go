package rl_agent

import "sync"

type Transition struct {
	State     SystemState
	Action    Action
	Reward    float64
	NextState SystemState
	Done      bool
}

type ReplayBuffer struct {
	buffer []Transition
	capacity int
	index    int
	mu      sync.Mutex
}

func NewReplayBuffer(capacity int) *ReplayBuffer {
	return &ReplayBuffer{
		buffer: make([]Transition, capacity),
		capacity: capacity,
	}
}

func (rb *ReplayBuffer) Push(t Transition) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buffer[rb.index] = t
	rb.index = (rb.index + 1) % rb.capacity
}

func (rb *ReplayBuffer) Sample(batchSize int) []Transition {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	// 简化的随机采样（实际应使用更高效的采样方式）
	return rb.buffer[:batchSize]
}