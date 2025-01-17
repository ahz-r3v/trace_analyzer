package wrapper

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"trace-analyser/pkg/info"
)

// ParseAndConvertCSV2 processes the second format of CSV file and converts it into invocation timestamps.
func ParseAndConvertAzure2021(inputFilePath string) ([]info.FunctionInvocations, error) {
	// Open the CSV file
	file, err := os.Open(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse the CSV file
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Define the results
	results := make(map[string]info.FunctionInvocations)

	// Process each row in the CSV
	for i, row := range rows {
		// Skip header row
		if i == 0 {
			continue
		}

		// Ensure row has expected number of columns
		if len(row) < 4 {
			return nil, fmt.Errorf("invalid row format at line %d", i+1)
		}

		app := row[0]
		function := row[1]
		hashFunction := app + function

		// Parse end timestamp (in seconds)
		endTimestamp, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end timestamp at line %d: %w", i+1, err)
		}

		// Parse duration (in seconds)
		duration, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid duration at line %d: %w", i+1, err)
		}

		// Convert timestamps and duration to milliseconds
		endTimeMillis := float64(endTimestamp * 1000) 
		durationMillis := float64(duration * 1000)

		// Calculate start timestamp
		startTimeMillis := endTimeMillis - durationMillis 

		// Update the results map
		if _, exists := results[hashFunction]; !exists {
			results[hashFunction] = info.FunctionInvocations{
				FunctionName: hashFunction,
				Timestamps:   []float64{},
				Durations:     []float64{},
			}
		}

		// Append the invocation data
		invocation := results[hashFunction]
		invocation.Timestamps = append(invocation.Timestamps, startTimeMillis)
		invocation.Durations = append(invocation.Durations, durationMillis)
		results[hashFunction] = invocation
	}

	// Convert map to slice
	var resultSlice []info.FunctionInvocations
	for _, invocation := range results {
		resultSlice = append(resultSlice, invocation)
	}

	return resultSlice, nil
}
