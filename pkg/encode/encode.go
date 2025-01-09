package encode

import (
	"encoding/csv"
	"log"
	"os"
	"sort"
	"strconv"
	"trace-analyser/pkg/info"
)

func EncodeToCSV(allColdstarts []info.LabeledTimestamp, coldstartsFrom0 []info.LabeledTimestamp, 
	  periodicColdstarts []info.LabeledTimestamp, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer file.Close()

	writer := csv.NewWriter(file)
    defer writer.Flush() 

	header := []string{"FunctionName", "Time", "ColdstartFrom0", "PeriodicInvocation"}
    if err := writer.Write(header); err != nil {
        log.Fatalf("Failed to write header: %v", err)
    }

	sort.Slice(allColdstarts, func(i, j int) bool {
        return allColdstarts[i].Timestamp < allColdstarts[j].Timestamp
    })
	sort.Slice(coldstartsFrom0, func(i, j int) bool {
        return coldstartsFrom0[i].Timestamp < coldstartsFrom0[j].Timestamp
    })
	sort.Slice(periodicColdstarts, func(i, j int) bool {
        return periodicColdstarts[i].Timestamp < periodicColdstarts[j].Timestamp
    })

    j, k := 0, 0
    // Multi-pointer linear merge
    for _, val := range allColdstarts {
        for j < len(coldstartsFrom0) && coldstartsFrom0[j].Timestamp < val.Timestamp {
            j++
        }
        from0 := "false"
        if j < len(coldstartsFrom0) && coldstartsFrom0[j].Timestamp == val.Timestamp {
            from0 = "true"
        }

        for k < len(periodicColdstarts) && periodicColdstarts[k].Timestamp < val.Timestamp {
            k++
        }
        periodic := "false"
        if k < len(periodicColdstarts) && periodicColdstarts[k].Timestamp == val.Timestamp {
            periodic = "true"
        }
		// output
		time := strconv.FormatFloat(val.Timestamp, 'f', -1, 64)
		functionName := val.FunctionName
		row := []string{functionName, time, from0, periodic}
		if err := writer.Write(row); err != nil {
			log.Fatalf("Failed to write row: %v", err)
		}
    }
	writer.Flush()

	if err := writer.Error(); err != nil {
        log.Fatalf("Error writing csv: %v", err)
    }
    log.Println("CSV file written successfully (line by line).")

	return nil
}