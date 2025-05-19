from flask import Flask, request, jsonify
from stable_baselines3 import PPO
import numpy as np
import torch
import os

app = Flask(__name__)

# Define environment simulation helpers
NUM_ORDERERS = 3
NUM_BUCKETS = 64  # default, can be dynamically overridden by input

# Load PPO model (must be pre-trained and saved)
MODEL_PATH = "ppo_bucket_assignment.zip"
if not os.path.exists(MODEL_PATH):
    raise FileNotFoundError("Trained PPO model not found. Please train and save model first.")

model = PPO.load(MODEL_PATH)

# Encode RLState (as received from Go) to numpy observation for PPO
# Here we simply flatten TxCount, CrossRatio, CPU, Latency
# Assumes consistent bucket & orderer count between training and inference
def encode_state(state):
    obs = []
    for b in sorted(state["bucket_stats"], key=lambda x: x["id"]):
        # obs.extend([
        obs.append([
            b["tx_count"] / 100.0,
            b["cross_ratio"]
        ])
    return np.array(obs, dtype=np.float32)

# Decode PPO action (array of orderer indices) to assignment map
# action[i] = j means bucket i -> orderer j
def decode_action(action, bucket_ids, orderer_ids):
    assignment = {oid: [] for oid in orderer_ids}
    for i, oid_index in enumerate(action):
        oid = orderer_ids[oid_index]
        assignment[oid].append(bucket_ids[i])
    return assignment

@app.route("/decide", methods=["POST"])
def decide():
    req_data = request.json
    state = req_data.get("state")
    if not state:
        return jsonify({"error": "Missing state field"}), 400

    print("Received state:")
    print(state)
    print("Bucket stats length:", len(state["bucket_stats"]))

    bucket_ids = [b["id"] for b in state["bucket_stats"]]
    orderer_ids = [o for o in state["orderers"]]

    obs = encode_state(state)
    action, _ = model.predict(obs, deterministic=True)

    assignment = decode_action(action, bucket_ids, orderer_ids)
    return jsonify({"assignment": assignment})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)