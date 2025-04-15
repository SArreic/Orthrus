package manager

import (
    "github.com/Hanzheng2021/orthrus/rl_agent"
    "github.com/Hanzheng2021/orthrus/config"
)

type RLDecorator struct {
    base   Manager
    agent  *rl_agent.DQNAgent
    cfg    *config.Config
}

func NewRLDecorator(base Manager, cfg *config.configuration) *RLDecorator {
    return &RLDecorator{
        base:  base,
        agent: rl_agent.NewDQNAgent(cfg.RL.StateDim, cfg.RL.ActionDim),
        cfg:   cfg,
    }
}

func (d *RLDecorator) Start() error {
    return d.base.Start()
}

func (d *RLDecorator) Stop() error {
    return d.base.Stop()
}

func (d *RLDecorator) GetLeaders() []NodeID {
    return d.base.GetLeaders()
}

func (d *RLDecorator) collectRLState() rl_agent.SystemState {
    stats := d.base.GetLoadStats()
    return rl_agent.SystemState{
        BucketLoads:  convertToFloat64(stats.BucketLoads),
        LeaderLoads:  stats.LeaderLoads,
        NetworkDelay: stats.NetworkDelay,
    }
}

func (d *RLDecorator) ApplyRLAction(action DynamicAction) error {
    if d.base == nil {
        return errors.New("base manager is nil")
    }
    return d.base.ApplyRLAction(action)
}

func (d *RLDecorator) ResizeLeaderGroup(delta int) error {
    return d.base.ResizeLeaderGroup(delta)
}