package info

// InvocationTimestamps holds the timestamps of all invocations for a specific function.
type FunctionInvocation struct {
	FunctionName	string    // Function identifier
	Timestamps   	[]float64 // List of invocation timestamps (in seconds since epoch)
	Duration     	[]float64 // List of invocation durations (in seconds since epoch)
}

type LabeledTimestamp struct {
	Timestamp       float64
	FunctionName    string
}