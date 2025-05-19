import gymnasium as gym
from gymnasium import spaces, Env
import numpy as np
import torch
from stable_baselines3 import PPO
from stable_baselines3.common.env_checker import check_env

class BucketAssignmentEnv(Env):
    def __init__(self, num_orderers=3, num_buckets=64):
        super().__init__()
        self.num_orderers = num_orderers
        self.num_buckets = num_buckets
        self.state = None
        self.action_space = spaces.MultiDiscrete([num_orderers] * num_buckets)
        self.observation_space = spaces.Box(low=0, high=1, shape=(num_buckets, 2), dtype=np.float32)

    def reset(self, *, seed=None, options=None):
        super().reset(seed=seed)
        self.state = np.random.rand(self.num_buckets, 2).astype(np.float32)
        return self.state, {}

    def step(self, action):
        reward = self._evaluate(action)
        terminated = True
        truncated = False
        return self.state, reward, terminated, truncated, {}

    def _evaluate(self, action):
        counts = [0] * self.num_orderers
        for oid in action:
            counts[oid] += 1
        imbalance = max(counts) - min(counts)
        cross_penalty = np.random.uniform(0, 1)
        return -imbalance - cross_penalty

if __name__ == '__main__':
    env = BucketAssignmentEnv(num_orderers=3, num_buckets=64)
    check_env(env, warn=True)
    model = PPO("MlpPolicy", env, verbose=1)
    model.learn(total_timesteps=20000)
    model.save("ppo_bucket_assignment")