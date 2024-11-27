package main

import (
	"fmt"
	"log"
	"trace-analyser/pkg/logic"
	"trace-analyser/pkg/plot"
	"trace-analyser/pkg/wrapper"
	"time"
)

func main() {
	// Path to the input CSV file
	traceFile := "data/azure-2019/invocations_per_function_md.anon.d01.csv"
	duraFile := "data/azure-2019/function_durations_percentiles.anon.d01.csv"
	// azure2021File := "data/azure-2021/azure2021.txt"

	// Step 1: Process the trace file and get invocation data
	startOfDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)) // Fixed day for calculation
	invocationTimestamps, err := wrapper.ParseAndConvertAzure2019(traceFile, duraFile, startOfDay)
	// invocationTimestamps, err := wrapper.ParseAndConvertAzure2021(azure2021File, startOfDay)
	if err != nil {
		log.Fatalf("Error processing trace file: %v", err)
	}

	// Step 2: Analyze cold starts for each function
	analyzer := logic.ColdStartAnalyzer{KeepAlive: uint64(60)}

	// allInvocations, _ := analyzer.ExpandInvocations(invocationTimestamps)

	periodicInvocations, nonPeriodicInvocations, err := analyzer.FilterPeriodicInvocations(invocationTimestamps, 100)
	if err != nil {
		log.Fatalf("Error finding Periodic invocations: %v", err)
	} 

	periodicColdStarts, err := analyzer.AnalyzeColdStarts(periodicInvocations)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	nonPeriodicColdStarts, err := analyzer.AnalyzeColdStarts(nonPeriodicInvocations)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	coldStartTimestamps, err := analyzer.AnalyzeColdStarts(invocationTimestamps)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	coldStartTimestampsFrom0, err := analyzer.AnalyzeColdStartsFrom0(invocationTimestamps)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	// Step 3: Plot cold start statistics
	fmt.Printf("len(clodStartTimestamps): %d\n", (len(coldStartTimestamps)))
	startOfDayMilliSec := uint64(startOfDay.UnixNano()) / 1e6 // Fixed day for calculation

	allData := [][]uint64{coldStartTimestamps, coldStartTimestampsFrom0}
	legends := []string{"cold starts from N", "cold starts from 0"}
	alphas := []float64{1, 1}
	err = plot.PlotMultipleColdStarts(allData, legends, alphas, startOfDayMilliSec, "both_cold_starts_1.png")
	if err != nil {
		log.Fatalf("Error creating plot: %v", err)
	}
	fmt.Printf("Cold start statistics plotted successfully: %s\n", "cold_starts_per_minute.png")

	allData = [][]uint64{coldStartTimestamps, periodicColdStarts, nonPeriodicColdStarts}
	legends = []string{"cold starts from N", "periodic Cold Starts from N", "non-periodic Cold Starts"}
	alphas = []float64{1, 1, 1}
	err = plot.PlotMultipleColdStarts(allData, legends, alphas, startOfDayMilliSec, "3_periodic_cold_starts_1.png")
	if err != nil {
		log.Fatalf("Error creating plot: %v", err)
	}
	fmt.Printf("Cold start statistics plotted successfully: %s\n", "cold_starts_per_minute.png")

	allData = [][]uint64{coldStartTimestamps, periodicColdStarts}
	legends = []string{"cold starts from N", "periodic Cold Starts from N"}
	alphas = []float64{1, 1}
	err = plot.PlotMultipleColdStarts(allData, legends, alphas, startOfDayMilliSec, "periodic_cold_starts_1.png")
	if err != nil {
		log.Fatalf("Error creating plot: %v", err)
	}
	fmt.Printf("Cold start statistics plotted successfully: %s\n", "cold_starts_per_minute.png")

	allData = [][]uint64{nonPeriodicColdStarts}
	legends = []string{"non-periodic Cold Starts"}
	alphas = []float64{1}
	err = plot.PlotMultipleColdStarts(allData, legends, alphas, startOfDayMilliSec, "non-periodic_cold_starts_1.png")
	if err != nil {
		log.Fatalf("Error creating plot: %v", err)
	}
	fmt.Printf("Cold start statistics plotted successfully: %s\n", "cold_starts_per_minute.png")
	
}
