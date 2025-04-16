package rl_agent

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const PythonAgentAddr = "127.0.0.1:5555"

func QueryRLAction(state State) (Action, error) {
	conn, err := net.DialTimeout("tcp", PythonAgentAddr, 2*time.Second)
	if err != nil {
		return Action{}, fmt.Errorf("connect to RL agent failed: %v", err)
	}
	defer conn.Close()

	data, _ := json.Marshal(state)
	conn.Write(data)

	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return Action{}, fmt.Errorf("read from RL agent failed: %v", err)
	}

	var act Action
	if err := json.Unmarshal(buf[:n], &act); err != nil {
		return Action{}, fmt.Errorf("invalid response: %v", err)
	}
	return act, nil
}
