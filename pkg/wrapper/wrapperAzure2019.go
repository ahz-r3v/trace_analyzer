package wrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	// "regexp"
	// "strings"
	"trace-analyser/pkg/info"
)


// ParseAndConvert reads the input CSV file, processes invocation counts, and converts them into timestamps.
func ParseAndConvertAzure2019(invocationFilePath string, durationFilePath string, startOfDay time.Time) ([]info.InvocationTimestamps, error) {
	// Open the CSV file
	invoFile, err := os.Open(invocationFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer invoFile.Close()
	duraFile, err := os.Open(durationFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer duraFile.Close()

	// Parse the CSV file
	invoReader := csv.NewReader(invoFile)
	invoRows, err := invoReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	duraReader := csv.NewReader(duraFile)
	duraRows, err := duraReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	var results []info.InvocationTimestamps
	// blocks := strings.Split(invocationFilePath, ".")
	// re := regexp.MustCompile(`d(\d+)`)
	// matches := re.FindStringSubmatch(blocks[len(blocks)-2])
	// day, _ := strconv.Atoi(matches[1])
	// startOfDay := time.Date(2019, 7, day, 0, 0, 0, 0, time.UTC) // Set a fixed date for calculation
	// startOfDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Set a fixed date for calculation

	// Process invocation
	for i, row := range invoRows {
		// Skip header row
		if i == 0 {
			continue
		}
		if i % 2500 == 0 {
			fmt.Printf("Progerss: %.2f\n", float32(i / len(invoRows)))
		}
		// hashApp := row[1]
		hashFunction := row[1] + row[2]

		// Search in durationFile to find the corresponding duration
		var durations []uint64
		var duration uint64
		for j, duraRow := range duraRows {
			if duraRow[1]+duraRow[2] == hashFunction {
				d, err := strconv.Atoi(duraRow[3])
				if err != nil {
					return nil, fmt.Errorf("invalid duration at line %d, column %d: %w", j+1, 3, err)
				}
				duration = uint64(d) // Convert duration to milliseconds
				break
			}
		}

		// Convert invocation counts to timestamps
		var timestamps []uint64
		for minute, countStr := range row[4:] {
			count, err := strconv.Atoi(countStr)
			if err != nil {
				return nil, fmt.Errorf("invalid invocation count at line %d, column %d: %w", i+1, minute+5, err)
			}

			if count > 0 {
				// Limiting for testing only; without this limit, it may cause OOM on a 16GB memory system.
				if count > 600 {
					count = 600
				}
				// Calculate timestamps for the current minute
				minuteStart := startOfDay.Add(time.Duration(minute) * time.Minute)
				interval := time.Second * 60 / time.Duration(count) // Interval in time.Duration

				for j := 0; j < count; j++ {
					invocationTime := minuteStart.Add(time.Duration(j) * interval)
					timestamps = append(timestamps, uint64(invocationTime.UnixNano()/1e6))
					durations = append(durations, duration)
				}
			}
		}


		results = append(results, info.InvocationTimestamps{
			// HashApp:      hashApp,
			HashFunction: hashFunction,
			Timestamps:   timestamps,
			Duration: 	  durations,
		})
	}
	log.Println("wrapper.ParseAndConvert return")
	return results, nil
}
