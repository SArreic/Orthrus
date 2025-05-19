package manager

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

type RLClient struct {
    Endpoint string // e.g. "http://127.0.0.1:5000/decide"
    Timeout  time.Duration
}

type BucketStat struct {
    ID         int     `json:"id"`
    TxCount    int     `json:"tx_count"`
    CrossRatio float64 `json:"cross_ratio"`
}

type OrdererStat struct {
    ID      int32   `json:"id"`
    CPU     float64 `json:"cpu"`
    Latency float64 `json:"latency"`
}

type RLState struct {
    Epoch        int32         `json:"epoch"`
    Orderers     []int32       `json:"orderers"`
    BucketStats  []BucketStat  `json:"bucket_stats"`
    OrdererStats []OrdererStat `json:"orderer_stats"`
}

type RLAction map[int32][]int

func (c *RLClient) DecideAssignment(state RLState) (RLAction, error) {
    body, err := json.Marshal(map[string]interface{}{"state": state})
    if err != nil {
        return nil, fmt.Errorf("marshal error: %v", err)
    }

    client := &http.Client{Timeout: c.Timeout}
    resp, err := client.Post(c.Endpoint, "application/json", bytes.NewBuffer(body))
    if err != nil {
        return nil, fmt.Errorf("post request error: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("RL server returned status %s", resp.Status)
    }

    var result struct {
        Assignment RLAction `json:"assignment"`
    }
    data, _ := ioutil.ReadAll(resp.Body)
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("unmarshal error: %v", err)
    }
    return result.Assignment, nil
}
