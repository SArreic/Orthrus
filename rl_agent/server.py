import numpy as np
import socket
import json
import random

from stable_baselines3 import PPO

model = PPO.load("model/ppo")

def decide_action(state):
    # 固定只取前 4 个 CPU 利用率和队列长度，不足则补 0
    cpu = (state["CPUUtilization"] + [0]*4)[:4]
    qlen = (state["QueueLengths"] + [0]*4)[:4]
    other = [state["CrossInstanceRatio"], state["AvgNetworkLatency"]]
    obs = np.array(cpu + qlen + other, dtype=np.float32)
    
    # 输出调试信息
    print(f"🧮 RL Input obs.shape = {obs.shape}, obs = {obs}")
    
    action, _ = model.predict(obs, deterministic=True)
    return {"Type": int(action)}

def run_server():
    s = socket.socket()
    s.bind(("0.0.0.0", 5555))
    s.listen(5)
    print("RL Agent server started at port 5555")

    while True:
        conn, addr = s.accept()
        with conn:
            try:
                data = conn.recv(2048)
                state = json.loads(data.decode())
                print("📥 Received state from Go:", state)

                action = decide_action(state)
                print("📤 Sending action:", action)
                conn.send(json.dumps(action).encode())

            except Exception as e:
                print("❌ Error during request:", e)
                fallback = {"Type": 2}
                conn.send(json.dumps(fallback).encode())

if __name__ == "__main__":
    run_server()
