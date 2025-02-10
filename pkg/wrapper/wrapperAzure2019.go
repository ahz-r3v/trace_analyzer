package wrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"trace-analyser/pkg/info"

	"github.com/gocarina/gocsv"
	"github.com/vhive-serverless/loader/pkg/common"
	"github.com/vhive-serverless/loader/pkg/generator"
)

// ParseAndConvert reads the input CSV file, processes invocation counts, and converts them into timestamps.
func ParseAndConvertAzure2019(
	invocationFilePath string,
	durationFilePath string,
	memoryFilePath string,
	iatDistribution common.IatDistribution, // common.Exponential / common.Uniform / common.Equidistant
	shiftIAT bool,
	granularity common.TraceGranularity,
) ([]info.FunctionInvocations, error) {
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

	memoryFile, err := os.Open(memoryFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer memoryFile.Close()

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

	functionMemoryStatsList := []*common.FunctionMemoryStats{}
	err = gocsv.UnmarshalFile(memoryFile, &functionMemoryStatsList)
	if err != nil { // Load all durations from file
		return nil, fmt.Errorf("failed parsing memory file: %w", err)
	}

	results := make([]info.FunctionInvocations, 0)
	for i, row := range invoRows {
		if i == 0 {
			continue
		}
		// Parse function invocation stats
		var invocations []int
		hashOwner := row[0]
		hashApp := row[1]
		hashFunction := row[2]
		index := hashOwner + hashApp + hashFunction
		memoryIndex := hashOwner + hashApp
		trigger := row[3]

		for j := 4; j < len(row); j++ {
			invocationCount, err := strconv.Atoi(row[j])
			if err != nil {
				return nil, fmt.Errorf("failed parsing invocation file, cannot convert %s to int: %w", row[j], err)
			}
			invocations = append(invocations, invocationCount)
		}

		var funcInvocationStats = common.FunctionInvocationStats{
			HashOwner:    hashOwner,
			HashApp:      hashApp,
			HashFunction: hashFunction,
			Trigger:      trigger,

			Invocations: invocations,
		}
		var function = common.Function{}
		// Search in durationFile to find the corresponding duration
		for i, functionRuntimeStats := range functionRuntimeStatsList {
			if functionRuntimeStats.HashOwner+functionRuntimeStats.HashApp+functionRuntimeStats.HashFunction == index {
				function.Name = index
				function.InvocationStats = &funcInvocationStats
				// fmt.Println(functionRuntimeStats)
				function.RuntimeStats = functionRuntimeStats
				// function.MemoryStats = &common.FunctionMemoryStats{}
				// function.MemoryStats = &common.FunctionMemoryStats{}
				break
			}
			if i == len(functionRuntimeStatsList) - 1 {
				fmt.Println("runtime not found!")
				function.Name = index
				function.InvocationStats = &funcInvocationStats
				function.RuntimeStats = functionRuntimeStats
			}
		}

		// Search in memoryFile to find the corresponding memory data
		for i, functionMemoryStats := range functionMemoryStatsList {
			if functionMemoryStats.HashOwner+functionMemoryStats.HashApp == memoryIndex {
				// fmt.Println(functionMemoryStats)
				function.MemoryStats = functionMemoryStats
				break
			}
			if i == len(functionMemoryStatsList) - 1 {
				fmt.Println("memory not found! " + memoryIndex)
				function.MemoryStats = functionMemoryStats
			}
		}

		// Convert invocation counts to timestamps
		var timestamps []float64
		var seed int64 = 123456789
		specGen := generator.NewSpecificationGenerator(seed)
		// fmt.Println(function.MemoryStats)
		// fmt.Println(function.RuntimeStats)
		if (function.RuntimeStats == nil) {
			panic("function runtime nil.")
		}
		if (function.MemoryStats == nil) {
			panic("function memory nil.")
		}
		specResult := specGen.GenerateInvocationData(&function, iatDistribution, shiftIAT, granularity)
		timestamps = expandByColumn(specResult.IAT)

		results = append(results, info.FunctionInvocations{
			// HashApp:      hashApp,
			FunctionName: index,
			Timestamps:   timestamps,
			Durations:    specResult.RawDuration,
		})
	}
	log.Println("wrapper.ParseAndConvert return")
	return results, nil
}

func expandByColumn(matrix [][]float64) []float64 {
    rows := len(matrix)
    if rows == 0 {
        return nil
    }
    maxCols := 0
    for _, row := range matrix {
        if len(row) > maxCols {
            maxCols = len(row)
        }
    }
    if maxCols == 0 {
        return nil
    }
    result := make([]float64, 0, rows*maxCols)
    for col := 0; col < maxCols; col++ {
        for row := 0; row < rows; row++ {
            if col < len(matrix[row]) {
                result = append(result, matrix[row][col])
            }
        }
    }
    return result
}

