from flask import Flask, request, jsonify
import random

app = Flask(__name__)

# For simplicity, we use a random strategy initially
# def dummy_decide_assignment(state):
#     orderers = state['orderers']
#     buckets = [b['id'] for b in state['bucket_stats']]
#     assignment = {oid: [] for oid in orderers}
#     for b in buckets:
#         oid = random.choice(orderers)
#         assignment[oid].append(b)
#     return assignment

def dummy_decide_assignment(state):
    orderers = state['orderers']
    buckets = [b['id'] for b in state['bucket_stats']]
    assignment = {}
    for i, b in enumerate(buckets):
        oid = orderers[i % len(orderers)]  # Round-robin
        assignment.setdefault(oid, []).append(b)
    return assignment

@app.route("/decide", methods=["POST"])
def decide():
    state = request.json.get("state")
    if state is None:
        return jsonify({"error": "Invalid input"}), 400
    result = dummy_decide_assignment(state)
    return jsonify({"assignment": result})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)