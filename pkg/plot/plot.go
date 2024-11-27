package plot

import (
	"fmt"
	"os"
	"time"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// PlotColdStarts takes the cold start timestamps in milliseconds and generates a chart showing the count per minute.
func PlotColdStarts(coldStartTimestamps []uint64, startOfDay uint64, outputFilePath string) error {
	// Constants for time calculations
	const MillisecondsPerMinute = 60000

	// Count cold starts per minute
	countsPerMinute := make(map[uint64]int)
	for _, timestamp := range coldStartTimestamps {
		// Truncate timestamp to the nearest minute
		minute := (timestamp / MillisecondsPerMinute) * MillisecondsPerMinute
		countsPerMinute[minute]++
	}

	// Generate x-axis labels and y-axis counts
	var xLabels []float64
	var yCounts []float64

	// Assume a full day (1440 minutes)
	total := 0.0
	for i := 0; i < 1440; i++ {
		minute := startOfDay + uint64(i)*MillisecondsPerMinute
		xLabels = append(xLabels, float64(minute))
		yCounts = append(yCounts, float64(countsPerMinute[minute]))
		total += float64(countsPerMinute[minute])
	}


	// Create the chart
	c := chart.Chart{
		Width:  1280,
		Height: 720,
		XAxis: chart.XAxis{
			Name: "Time (Minutes)",
			ValueFormatter: func(v interface{}) string {
				// Format the x-axis values as time strings
				ms := uint64(v.(float64))
				return FormatMillisecondsToTime(ms).Format("15:04")
			},
		},
		YAxis: chart.YAxis{
			Name: "Cold Starts",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: xLabels,
				YValues: yCounts,
			},
		},
	}

	// Write the chart to an output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Render the chart as PNG
	err = c.Render(chart.PNG, file)
	if err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	fmt.Printf("Chart saved to %s\n", outputFilePath)
	return nil
}

// FormatMillisecondsToTime converts a millisecond-level timestamp to time.Time for formatting purposes.
func FormatMillisecondsToTime(ms uint64) time.Time {
	return time.Unix(0, int64(ms)*1e6)
}

func getVividColor(i int, alpha uint8) drawing.Color {
	switch i {
	case 0: // First group: Vivid blue
		return drawing.Color{R: 0, G: 102, B: 204, A: alpha} // DeepBlue
	case 1: // Second group: Vivid green
		return drawing.Color{R: 204, G: 51, B: 51, A: alpha} // DeepGreen
	case 2: // Third group: Vivid red
		return drawing.Color{R: 0, G: 153, B: 76, A: alpha} // DeepRed
	default: // Default color (black) for unexpected cases
		return drawing.Color{R: 0, G: 0, B: 0, A: alpha} // Black
	}
}


func PlotMultipleColdStarts(
	coldStartTimestamps [][]uint64,
	legends []string,
	alphas []float64,
	startOfDay uint64,
	outputFilePath string,
) error {
	if len(coldStartTimestamps) != len(legends) || len(coldStartTimestamps) != len(alphas) {
		return fmt.Errorf("the number of data sets, legends, and alphas must match")
	}

	const MillisecondsPerMinute = 60000

	// Prepare data for each dataset
	var allSeries []chart.Series
	for i, timestamps := range coldStartTimestamps {
		// Count cold starts per minute for the current dataset
		countsPerMinute := make(map[uint64]int)
		for _, timestamp := range timestamps {
			minute := (timestamp / MillisecondsPerMinute) * MillisecondsPerMinute
			countsPerMinute[minute]++
		}

		// Generate x-axis labels and y-axis counts
		var xLabels []float64
		var yCounts []float64

		// Assume a full day (1440 minutes)
		total := 0.0
		for j := 0; j < 1440; j++ {
			minute := startOfDay + uint64(j)*MillisecondsPerMinute
			xLabels = append(xLabels, float64(minute))
			yCounts = append(yCounts, float64(countsPerMinute[minute]))
			total += float64(countsPerMinute[minute])
		}
		fmt.Println(legends[i] + " average:", total/1440)

		// Create a series for the current dataset with specified transparency and fixed colors
		alpha := uint8(alphas[i] * 255) // Convert transparency to 0-255 range
		allSeries = append(allSeries, chart.ContinuousSeries{
			Name:    fmt.Sprintf("%s Average: %.2f", legends[i], total/1440),
			XValues: xLabels,
			YValues: yCounts,
			Style: chart.Style{
				StrokeColor: getVividColor(i, alpha),
				StrokeWidth: 1.0,
			},
		})
	}

	// Enable legend
	c := chart.Chart{
		Width:  1280,
		Height: 720,
		XAxis: chart.XAxis{
			Name: "Time (Minutes)",
			ValueFormatter: func(v interface{}) string {
				ms := uint64(v.(float64))
				return FormatMillisecondsToTime(ms).Format("15:04")
			},
		},
		YAxis: chart.YAxis{
			Name: "Cold Starts",
		},
		Series: allSeries,
		Elements: []chart.Renderable{
			chart.LegendLeft(&chart.Chart{
				Series: allSeries,
			}),
		},
	}

	// Write the chart to an output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Render the chart as PNG
	err = c.Render(chart.PNG, file)
	if err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	fmt.Printf("Chart saved to %s\n", outputFilePath)
	return nil
}