// routing/metrics.go
package routing

type CrossInstanceRecorder interface {
	RecordCrossInstance(cross bool)
}

type StateCollector interface {
	CollectState(numInstances int) State
}

var (
	crossInstanceRecorder CrossInstanceRecorder
	stateCollector        StateCollector
)

func SetCrossInstanceRecorder(r CrossInstanceRecorder) {
	crossInstanceRecorder = r
}

func SetStateCollector(s StateCollector) {
	stateCollector = s
}

func RecordCrossInstance(cross bool) {
	if crossInstanceRecorder != nil {
		crossInstanceRecorder.RecordCrossInstance(cross)
	}
}

func CollectState(numInstances int) State {
	if stateCollector != nil {
		return stateCollector.CollectState(numInstances)
	}
	return State{} // fallback 空状态
}
