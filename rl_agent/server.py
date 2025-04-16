import numpy as np
import socket
import json
import random

from stable_baselines3 import PPO

model = PPO.load("rl_agent/model/ppo")

def decide_action(state):
    obs = np.array(state["CPUUtilization"] + state["QueueLengths"] + [state["CrossInstanceRatio"], state["AvgNetworkLatency"]])
    obs = obs.astype(np.float32)
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
