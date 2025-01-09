package info

// InvocationTimestamps holds the timestamps of all invocations for a specific function.
type FunctionInvocations struct {
	FunctionName    string    // Function identifier
	Timestamps      []float64 // List of invocation timestamps (in seconds)
	Durations       []float64 // List of invocation durations (in seconds)
}

type LabeledTimestamp struct {
	Timestamp       float64
	FunctionName    string
}