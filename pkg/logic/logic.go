package logic

import (
	// "time"
	"fmt"
	"trace-analyser/pkg/info"
	"log"
)

// ColdStartAnalyzer analyzes invocation data and determines cold start timestamps.
type ColdStartAnalyzer struct {
	KeepAlive float64 // Time (e.g., 60 seconds) an instance remains alive after invocation ends
}

type Instance struct {
	LastEndTime float64
	ExpiryTime float64
}

// FilterPeriodicInvocations identifies periodic invocation patterns across all functions.
// Returns two slices: one for periodic invocations and another for non-periodic invocations.
func (c *ColdStartAnalyzer) FilterPeriodicInvocations(
	invocations []info.FunctionInvocation,
	tolerance float64,
) ([]info.FunctionInvocation, []info.FunctionInvocation, error) {
	const minPeriodicDuration float64 = 12 * 60 * 60 * 1000 // 12 hours in milliseconds

	var periodicInvocations []info.FunctionInvocation
	var nonPeriodicInvocations []info.FunctionInvocation

	for _, invocationData := range invocations {
		timestamps := invocationData.Timestamps
		if len(timestamps) < 2 {
			// Less than 2 invocations; cannot determine periodicity
			nonPeriodicInvocations = append(nonPeriodicInvocations, invocationData)
			continue
		}

		// Calculate intervals between consecutive invocations
		intervalCounts := make(map[float64]int)
		for i := 1; i < len(timestamps); i++ {
			interval := timestamps[i] - timestamps[i-1]

			// Group intervals within the tolerance range
			foundMatch := false
			for existingInterval := range intervalCounts {
				if absDiff(interval, existingInterval) <= tolerance {
					intervalCounts[existingInterval]++
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				intervalCounts[interval]++
			}
		}

		// Find the most common interval
		var maxFrequency int
		var mostFrequentInterval float64
		for interval, frequency := range intervalCounts {
			if frequency > maxFrequency {
				maxFrequency = frequency
				mostFrequentInterval = interval
			}
		}

		// Verify if the invocation is periodic
		totalPeriodicDuration := mostFrequentInterval * float64(maxFrequency)
		if totalPeriodicDuration >= minPeriodicDuration {
			// Add to periodic invocations
			periodicInvocations = append(periodicInvocations, invocationData)
		} else {
			// Add to non-periodic invocations
			nonPeriodicInvocations = append(nonPeriodicInvocations, invocationData)
		}
	}

	return periodicInvocations, nonPeriodicInvocations, nil
}

// absDiff calculates the absolute difference between two uint64 numbers.
func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// AnalyzeColdStarts processes invocation timestamps and durations for multiple functions,
// returning a map where the key is the function identifier and the value is a list of cold start timestamps.
func (c *ColdStartAnalyzer) AnalyzeColdStarts(invocations []info.FunctionInvocation) ([]info.LabeledTimestamp, error) {
	results := make([]info.LabeledTimestamp, 0)

	for _, invocationData := range invocations {
		timestamps := invocationData.Timestamps
		durations := invocationData.Duration
		functionName := invocationData.FunctionName

		// Check slice length
		if len(timestamps) != len(durations) {
			return nil, fmt.Errorf("len(timestamps) is %d while len(durations) is %d", len(timestamps), len(durations))
		}

		// Track active instances and their expiry times for this specific function
		activeInstances := make([]Instance, 0) // Key: instance last end time, Value: expiry time
		var coldStartTimestamps []info.LabeledTimestamp

		for i, start := range timestamps {
			duration := durations[i]
			instanceFound := false
			// fmt.Println("s:", start)
			// fmt.Println(duration)

			// Check if any instance is available
			// for i, instance := range activeInstances {
			for j := len(activeInstances) - 1; j >= 0; j-- {
				instance := activeInstances[j]
				if start > instance.LastEndTime && start < instance.ExpiryTime {
					// Use this instance and update its expiry time
					// Update instance info
					activeInstances[j].LastEndTime = start + duration
					activeInstances[j].ExpiryTime = start + duration + c.KeepAlive*1000
					instanceFound = true
					break
				}
			}

			// If no active instance found, this is a cold start
			if !instanceFound {
				coldStartTimestamps = append(
					coldStartTimestamps,
					info.LabeledTimestamp{
						Timestamp:    start,
						FunctionName: functionName,
					},
			    )
				// Create a new instance and set its expiry time
				activeInstances = append(activeInstances, Instance{
					LastEndTime: start + duration,
					ExpiryTime:  start + duration + c.KeepAlive*1000,
				})
			}

			for j := len(activeInstances) - 1; j >= 0; j-- {
				instance := activeInstances[j]
				if start > instance.ExpiryTime {
					activeInstances = append(activeInstances[:j], activeInstances[j+1:]...)
				}
			}

			instanceFound = false
		}

		// Store cold start timestamps for this specific function
		results = append(results, coldStartTimestamps...)
	}

	log.Println("logic.AnalyzeColdStarts return")
	return results, nil
}

// AnalyzeColdStartsFrom0 calculates cold starts from 0 instance.
func (c *ColdStartAnalyzer) AnalyzeColdStartsFrom0(invocations []info.FunctionInvocation) ([]info.LabeledTimestamp, error) {
	results := make([]info.LabeledTimestamp, 0)

	for _, invocationData := range invocations {
		timestamps := invocationData.Timestamps
		durations := invocationData.Duration
		functionName := invocationData.FunctionName

		// Check slice length
		if len(timestamps) != len(durations) {
			return nil, fmt.Errorf("len(timestamps) is %d while len(durations) is %d", len(timestamps), len(durations))
		}

		// Track active instances and their expiry times for this specific function
		activeInstances := make([]Instance, 0)
		var coldStartTimestamps []info.LabeledTimestamp

		for i, start := range timestamps {
			duration := durations[i]
			instanceFound := false

			// Check if any instance is available
			for j := len(activeInstances) - 1; j >= 0; j-- {
				instance := activeInstances[j]
				if start <= instance.ExpiryTime {
					// Use this instance and update its expiry time
					activeInstances[j].ExpiryTime = start + c.KeepAlive*1000 + duration
					instanceFound = true
					break
				} else if start > instance.ExpiryTime {
					// Remove expired instance
					activeInstances = append(activeInstances[:j], activeInstances[j+1:]...)
				}
			}

			// If no active instance is found, this is a cold start
			if !instanceFound {
				coldStartTimestamps = append(
					coldStartTimestamps, 
					info.LabeledTimestamp{
						Timestamp:    start,
						FunctionName: functionName,
					},
				)
				// Create a new instance and set its expiry time
				activeInstances = append(activeInstances, Instance{
					LastEndTime: start,
					ExpiryTime:  start + c.KeepAlive*1000 + duration,
				})
			}
		}

		// Store cold start timestamps for this specific function
		results = append(results, coldStartTimestamps...)
	}

	log.Println("logic.AnalyzeColdStartsFrom0 return")
	return results, nil
}

func (c *ColdStartAnalyzer) ExpandInvocations(invocations []info.FunctionInvocation) ([]float64, error) {
	var allStartTimestamps []float64

	for _, invocation := range invocations {
		// 检查 Timestamps 是否为空
		if len(invocation.Timestamps) == 0 {
			continue
		}

		allStartTimestamps = append(allStartTimestamps, invocation.Timestamps...)
	}

	return allStartTimestamps, nil
}