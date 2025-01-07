package wrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"trace-analyser/pkg/info"
	"github.com/vhive-serverless/loader/pkg/generator"
	"github.com/vhive-serverless/loader/pkg/common"
	"github.com/gocarina/gocsv"
)


// ParseAndConvert reads the input CSV file, processes invocation counts, and converts them into timestamps.
func ParseAndConvertAzure2019(
	  invocationFilePath string, 
	  durationFilePath string, 
	  startOfDay time.Time,
	  iatDistribution common.IatDistribution,// common.Exponential / common.Uniform / common.Equidistant
	  shiftIAT bool,
	  granularity common.TraceGranularity,
	) ([]info.FunctionInvocation, error) {
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

	functionRuntimeStatsList := []*common.FunctionRuntimeStats{}
	err = gocsv.UnmarshalFile(duraFile, &functionRuntimeStatsList)
	if err != nil { // Load all durations from file
		return nil, fmt.Errorf("failed parsing duration file: %w", err)
	}

	results := make([]info.FunctionInvocation, 0)
	for i, row := range invoRows {
		if i == 0 {
			continue
		}
		// Parse function invocation stats
		var invocations []int
		hashOwner := row[0]
		hashApp := row[1]
		hashFunction := row[2]
		index := hashOwner+ hashApp + hashFunction
		trigger := row[3]
		
		for j := 4; j < len(row); j++ {
			invocationCount, err := strconv.Atoi(row[j])
			if err != nil {
				return nil, fmt.Errorf("failed parsing invocation file, cannot convert %s to int: %w", row[j], err)
			}
			invocations = append(invocations, invocationCount)
		}

		var funcInvocationStats = common.FunctionInvocationStats{
			HashOwner: 		hashOwner,
			HashApp: 		hashApp,
			HashFunction: 	hashFunction,
			Trigger:		trigger,

			Invocations:	invocations,
		}
		var function = common.Function{}
		// Search in durationFile to find the corresponding duration
		for _, functionRuntimeStats := range functionRuntimeStatsList {
			if functionRuntimeStats.HashOwner + functionRuntimeStats.HashApp + functionRuntimeStats.HashFunction == index {
				function.Name =	index
				function.InvocationStats = &funcInvocationStats
				function.RuntimeStats =	functionRuntimeStats
				break
			} 
		}

		// Convert invocation counts to timestamps
		var timestamps []float64
		var seed int64 = 123456789
		specGen := generator.NewSpecificationGenerator(seed)
		specResult := specGen.GenerateInvocationData(&function, iatDistribution, shiftIAT, granularity)
		timestamps = expandByColumn(specResult.IAT)

		results = append(results, info.FunctionInvocation{
			// HashApp:      hashApp,
			FunctionName: index,
			Timestamps:   timestamps,
			Duration: 	  specResult.RawDuration,
		})
	}
	log.Println("wrapper.ParseAndConvert return")
	return results, nil
}

func expandByColumn(matrix [][]float64) []float64 {
	rows := len(matrix)
	cols := len(matrix[0])
	var result []float64
	for col := 0; col < cols; col++ { 
		for row := 0; row < rows; row++ {
			result = append(result, matrix[row][col])
		}
	}
	return result
}
