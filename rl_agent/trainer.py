import gymnasium as gym
from gymnasium import spaces
import numpy as np
from stable_baselines3 import PPO
from stable_baselines3.common.env_checker import check_env
import random

# ========== 自定义 Gym 环境：模拟 Orthrus 状态 ========== #
class OrthrusEnv(gym.Env):
    def __init__(self, num_instances=4):
        super(OrthrusEnv, self).__init__()
        self.num_instances = num_instances
        self.max_steps = 2000
        self.max_queue = 100

        # 状态空间：CPU、队列长度、跨实例率、延迟
        self.observation_space = spaces.Box(low=0, high=1, shape=(num_instances*2 + 2,), dtype=np.float32)

        # 动作空间：0 加实例、1 减实例、2 重新分配
        self.action_space = spaces.Discrete(3)

        self.state = self._mock_state()

    def _mock_state(self):
        cpu = [random.random() for _ in range(self.num_instances)]
        queue = [random.random() for _ in range(self.num_instances)]
        cross = random.random() * 0.3
        latency = random.random()
        return np.array(cpu + queue + [cross, latency], dtype=np.float32)

    def reset(self, *, seed=None, options=None):
        super().reset(seed=seed)
        self.state = self._mock_state()
        self.current_step = 0  # 每次 reset 时重置步数
        return self.state, {}


    def step(self, action):
        self.state = self._mock_state()
        TPS_delta = random.uniform(0, 1)
        cross = self.state[-2]
        latency = self.state[-1]
        reward = float(0.4 * TPS_delta - 0.3 * cross + 0.3 * (1.0 / np.log(latency + 1.1)))

        self.current_step += 1
        terminated = False  # 无终止状态
        truncated = self.current_step >= self.max_steps

        return self.state, reward, terminated, truncated, {}


# ========== 训练 PPO 模型 ========== #
if __name__ == "__main__":
    env = OrthrusEnv(num_instances=4)
    check_env(env)

    model = PPO("MlpPolicy", env, verbose=1)
    model.learn(total_timesteps=50000)

    model.save("rl_agent/model/ppo")
    print("✅ 模型训练完成并已保存")
