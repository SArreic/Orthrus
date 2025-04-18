// routing/latency.go
package routing

type LatencyRecorder interface {
	RecordLatency(ms int64)
}

var latencyRecorder LatencyRecorder

func SetLatencyRecorder(lr LatencyRecorder) {
	latencyRecorder = lr
}

func RecordLatency(ms int64) {
	if latencyRecorder != nil {
		latencyRecorder.RecordLatency(ms)
	}
}
