package runtime

// Runtime exposes runtime specific for all consumers
type Runtime interface {
	MetricsRecorder() Metrics
}

type DefaultRuntimeImpl struct {
	metricsRecorder Metrics
}

func NewRuntime() Runtime {
	return &DefaultRuntimeImpl{
		metricsRecorder: newMetricsImpl(),
	}
}

func (r *DefaultRuntimeImpl) MetricsRecorder() Metrics {
	return r.metricsRecorder
}
