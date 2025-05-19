import gym
import numpy as np
import torch
from stable_baselines3 import PPO
from stable_baselines3.common.env_checker import check_env
from gym import spaces

# Define a custom environment for bucket-to-orderer assignment
class BucketAssignmentEnv(gym.Env):
    def __init__(self, num_orderers=3, num_buckets=8):
        super().__init__()
        self.num_orderers = num_orderers
        self.num_buckets = num_buckets
        self.state = None
        self.action_space = spaces.MultiDiscrete([num_orderers] * num_buckets)  # one action per bucket
        self.observation_space = spaces.Box(low=0, high=1, shape=(num_buckets, 2), dtype=np.float32)  # dummy

    def reset(self):
        self.state = np.random.rand(self.num_buckets, 2)
        return self.state

    def step(self, action):
        reward = self._evaluate(action)
        done = True  # one-shot decision
        return self.state, reward, done, {}

    def _evaluate(self, action):
        counts = [0] * self.num_orderers
        for oid in action:
            counts[oid] += 1
        imbalance = max(counts) - min(counts)
        cross_penalty = np.random.uniform(0, 1)  # simulate cross-instance rate
        return -imbalance - cross_penalty  # reward = negative cost

# Train PPO model
if __name__ == '__main__':
    env = BucketAssignmentEnv()
    check_env(env, warn=True)
    model = PPO("MlpPolicy", env, verbose=1)
    model.learn(total_timesteps=20000)
    model.save("ppo_bucket_assignment")
