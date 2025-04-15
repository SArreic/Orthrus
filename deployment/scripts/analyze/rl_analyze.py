import pandas as pd
def plot_training_curve(log_path):
    df = pd.read_csv(log_path)
    df.plot(x='episode', y=['reward', 'cross_instance_rate'])