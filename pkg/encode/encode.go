package encode

import (
	"sort"
	"os"
	"log"
	"strconv"
	"encoding/csv"
)

func EncodeToCSV(allColdstarts []float64, coldstartsFrom0 []float64, periodicColdstarts []float64, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
        log.Fatalf("Failed to create file: %v", err)
    }
    defer file.Close()

	writer := csv.NewWriter(file)
    defer writer.Flush() 

	header := []string{"Tame", "ColdstartFrom0", "PeriodicInvocation"}
    if err := writer.Write(header); err != nil {
        log.Fatalf("Failed to write header: %v", err)
    }

	sort.Slice(allColdstarts, func(i, j int) bool {
        return allColdstarts[i] < allColdstarts[j]
    })
	sort.Slice(coldstartsFrom0, func(i, j int) bool {
        return coldstartsFrom0[i] < coldstartsFrom0[j]
    })
	sort.Slice(periodicColdstarts, func(i, j int) bool {
        return periodicColdstarts[i] < periodicColdstarts[j]
    })

    j, k := 0, 0
    // Multi-pointer linear merge
    for _, val := range allColdstarts {
        for j < len(coldstartsFrom0) && coldstartsFrom0[j] < val {
            j++
        }
        from0 := "false"
        if j < len(coldstartsFrom0) && coldstartsFrom0[j] == val {
            from0 = "true"
        }

        // 同理，向前移动 k，使 s3[k] >= val 或 k 超出边界
        for k < len(periodicColdstarts) && periodicColdstarts[k] < val {
            k++
        }
        periodic := "false"
        if k < len(periodicColdstarts) && periodicColdstarts[k] == val {
            periodic = "true"
        }
		// output
		time := strconv.FormatFloat(val, 'f', -1, 64)
		row := []string{time, from0, periodic}
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