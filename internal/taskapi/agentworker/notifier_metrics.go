package agentworker

// NotifierMetrics records harness notifier back-pressure drops. Implemented by
// taskapi worker metrics registration in production; tests may omit.
type NotifierMetrics interface {
	RecordNotifierDropped(kind string)
}
