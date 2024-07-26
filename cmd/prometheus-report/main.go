package main

import (
	"flag"
	"time"

	"prometheus-report-generator/config"
	"prometheus-report-generator/internal/query"
	"prometheus-report-generator/internal/report"
	"prometheus-report-generator/pkg/client"
	"prometheus-report-generator/pkg/logger"

	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

func main() {
	// Define command-line flags
	reportType := flag.String("reportType", "daily", "Type of the report: hourly, daily, weekly, monthly, quarterly")
	flag.Parse()

	// Load configuration
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		logrus.Fatalf("Error loading config: %v\n", err)
	}

	loggers := logger.InitLogger(
		config.Logging.Path,
		config.Logging.MaxAge,
		config.Logging.Compress,
	)

	// Determine time range and window size based on report type
	endTime := time.Now().UTC()
	var startTime time.Time
	var step time.Duration
	var window string
	switch *reportType {
	case "hourly":
		startTime = endTime.Add(-1 * time.Hour)
		step = 30 * time.Second
		window = "30s"
	case "daily":
		startTime = endTime.Add(-24 * time.Hour)
		step = 5 * time.Minute
		window = "5m"
	case "weekly":
		startTime = endTime.Add(-7 * 24 * time.Hour)
		step = 1 * time.Hour
		window = "1h"
	case "monthly":
		startTime = time.Date(endTime.Year(), endTime.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		step = 24 * time.Hour
		window = "1d"
	case "quarterly":
		quarterStartMonth := ((endTime.Month()-1)/3)*3 + 1
		year := endTime.Year()
		if quarterStartMonth <= 1 {
			year--
			quarterStartMonth = 10
		}
		startTime = time.Date(year, time.Month(quarterStartMonth), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -3, 0)
		step = 7 * 24 * time.Hour
		window = "1w"
	default:
		logrus.Fatalf("Invalid report type: %s", *reportType)
	}

	// Prepare to collect query results from all Prometheus instances
	queries := query.GetQueries(window)
	combinedResults := make(map[string]map[string]model.Value)

	// Query each Prometheus instance
	for platform, instanceConfig := range config.Prometheus {
		// Create Prometheus client
		promClient, err := client.CreatePrometheusClient(instanceConfig.URL)
		if err != nil {
			logger.Log(loggers, logrus.ErrorLevel, "Error creating Prometheus client for "+platform+" ("+instanceConfig.URL+"): "+err.Error())
			continue
		}

		// Run queries for each Prometheus instance
		for name, query := range queries {
			result, err := client.QueryRange(promClient, query, startTime, endTime, step)
			if err != nil {
				logger.Log(loggers, logrus.ErrorLevel, "Error querying Prometheus for "+name+" at "+platform+" ("+instanceConfig.URL+"): "+err.Error())
				continue
			}

			if combinedResults[platform] == nil {
				combinedResults[platform] = make(map[string]model.Value)
			}
			combinedResults[platform][name] = result
		}
	}

	// Generate report using combined results
	report.GenerateReport(combinedResults, loggers, *reportType)
}
