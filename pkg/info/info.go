package info

// InvocationTimestamps holds the timestamps of all invocations for a specific function.
type InvocationTimestamps struct {
	// HashApp      string    // Application identifier
	HashFunction string    // Function identifier
	Timestamps   []uint64 // List of invocation timestamps (in seconds since epoch)
	Duration     []uint64  // List of invocation durations (in seconds since epoch)
}