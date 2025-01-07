package main

import (
	"flag"
	// "fmt"
	"log"
	"time"
	// "sort"
	"trace-analyser/pkg/logic"
	"trace-analyser/pkg/info"
	"trace-analyser/pkg/wrapper"
	"trace-analyser/pkg/encode"
	"github.com/vhive-serverless/loader/pkg/common"
	// "encoding/csv"
)

// func main() {

// }

func main() {

	// Parse args
	wrapperType := flag.String("wrapper", "", "[azure2019 / azure2021]")
	keepAlive := flag.Float64("keepalive", 60, "Seconds an instance remains alive after invocation ends")
	tolerance := flag.Float64("tolerance", 100, "Tolerance (in milliseconds) for grouping intervals")
	iatDistribution := flag.Int("iatDistribution", 0, "[0=Exponential / 1=Uniform / 2=Equidistant]")
	shiftIAT := flag.Bool("shiftIAT", false, "shiftIAT")
	granularity := flag.Int("granularity", 0, "[0=MinuteGranularity / 1=SecondGranularity]")
	flag.Parse()
	nonFlagArgs := flag.Args()
	invocationFile := ""
	durationFile := ""
	outputPath := ""
	startOfDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60)) // Fixed day for calculation
	invocationTimestamps := make([]info.FunctionInvocation, 0)
	var err error

	switch *wrapperType{
	case "":
		log.Fatalf("-wrapper not defined.")
	case "azure2019":
		if len(nonFlagArgs) != 3 {
			log.Fatalf("-wrapper=azure2019 need 3 Args!")
		} else {
			invocationFile = nonFlagArgs[0]
			durationFile = nonFlagArgs[1]
			outputPath = nonFlagArgs[2]
			// Step 1: Process the trace file and get invocation data
			invocationTimestamps, err = wrapper.ParseAndConvertAzure2019(invocationFile, durationFile, startOfDay, 
				common.IatDistribution(*iatDistribution), *shiftIAT, common.TraceGranularity(*granularity))
			if err != nil {
				log.Fatalf("Error processing trace file: %v", err)
			}
		}
	case "azure2021":
		if len(nonFlagArgs) != 2 {
			log.Fatalf("-wrapper=azure2021 need 2 Args!")
		} else {
			invocationFile = nonFlagArgs[0]
			outputPath = nonFlagArgs[1]
			// Step 1: Process the trace file and get invocation data
			invocationTimestamps, err = wrapper.ParseAndConvertAzure2021(invocationFile, startOfDay)
			if err != nil {
				log.Fatalf("Error processing trace file: %v", err)
			}
		}
	default:
	}

	// // Step 1: Process the trace file and get invocation data
	// invocationTimestamps, err := wrapper.ParseAndConvertAzure2019(traceFile, duraFile, startOfDay)
	// // invocationTimestamps, err := wrapper.ParseAndConvertAzure2021(azure2021File, startOfDay)
	// if err != nil {
	// 	log.Fatalf("Error processing trace file: %v", err)
	// }

	// Step 2: Analyze cold starts for each function
	analyzer := logic.ColdStartAnalyzer{KeepAlive: float64(*keepAlive)}

	// allInvocations, _ := analyzer.ExpandInvocations(invocationTimestamps)

	periodicInvocations, _, err := analyzer.FilterPeriodicInvocations(invocationTimestamps, *tolerance)
	if err != nil {
		log.Fatalf("Error finding Periodic invocations: %v", err)
	} 

	periodicColdStarts, err := analyzer.AnalyzeColdStarts(periodicInvocations)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	coldStartTimestamps, err := analyzer.AnalyzeColdStarts(invocationTimestamps)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	coldStartFrom0Timestamps, err := analyzer.AnalyzeColdStartsFrom0(invocationTimestamps)
	if err != nil {
		log.Fatalf("Error calculating coldstarts: %v", err)
	}

	// Step 3: Output cold start statistics
	encode.EncodeToCSV(coldStartTimestamps, coldStartFrom0Timestamps, periodicColdStarts, outputPath)
}
