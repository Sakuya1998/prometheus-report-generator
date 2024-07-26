package report

import (
	"fmt"
	"time"

	"prometheus-report-generator/pkg/logger"

	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

func GenerateReport(queryResults map[string]map[string]model.Value, loggers map[logrus.Level]*logrus.Logger, reportType string) {
	const sheetName = "Report"
	headers := []string{"Platform", "Host", "Name", "CPU Max (%)", "CPU Min (%)", "CPU Avg (%)", "Memory Max (%)", "Memory Min (%)", "Memory Avg (%)", "Network Max (MB/s)", "Network Min (MB/s)", "Network Avg (MB/s)"}

	// Add Disk usage headers dynamically based on mount points
	mountPoints := getDiskMountPoints(queryResults)
	for _, mountPoint := range mountPoints {
		headers = append(headers, fmt.Sprintf("Disk %s Max (%%)", mountPoint), fmt.Sprintf("Disk %s Min (%%)", mountPoint), fmt.Sprintf("Disk %s Avg (%%)", mountPoint))
	}

	// Create a new Excel file
	file := excelize.NewFile()
	index, err := file.NewSheet(sheetName)
	if err != nil {
		logger.Log(loggers, logrus.ErrorLevel, fmt.Sprintf("Error creating new sheet: %v", err))
		return
	}

	// Remove the default Sheet1
	file.DeleteSheet("Sheet1")

	file.SetActiveSheet(index)

	// 设置标题样式
	headerStyle, err := file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})
	if err != nil {
		logger.Log(loggers, logrus.ErrorLevel, fmt.Sprintf("Error creating header style: %v", err))
		return
	}

	// 设置数据单元格样式
	dataStyle, err := file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		logger.Log(loggers, logrus.ErrorLevel, fmt.Sprintf("Error creating data style: %v", err))
		return
	}

	// Write headers
	for colIndex, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1)
		file.SetCellValue(sheetName, cell, header)
		file.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	rowIndex := 2
	for platform, platformResults := range queryResults {
		for host, name := range extractHostsAndNames(platformResults) {
			file.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), platform)
			file.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), host)
			file.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), name)

			file.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIndex), fmt.Sprintf("C%d", rowIndex), dataStyle)

			colIndex := 4
			for _, metric := range listMetrics() {
				max, min, avg := calculateStats(platformResults, host, metric)

				writeMetricValues(file, sheetName, rowIndex, colIndex, metric, max, min, avg, dataStyle)

				colIndex += 3
			}
			// Add disk usage values based on mount points
			for _, mountPoint := range mountPoints {
				max, min, avg := calculateDiskStats(platformResults, host, mountPoint)
				writeDiskValues(file, sheetName, rowIndex, colIndex, max, min, avg, dataStyle)
				colIndex += 3
			}
			rowIndex++
		}
	}

	// 调整列宽为自适应
	adjustColumnWidths(file, sheetName, len(headers), rowIndex-1)

	reportFilePath := fmt.Sprintf("./report_%s_%s.xlsx", reportType, time.Now().Format("20060102_150405"))
	if err := file.SaveAs(reportFilePath); err != nil {
		logger.Log(loggers, logrus.ErrorLevel, fmt.Sprintf("Error saving report file: %v", err))
		return
	}

	logger.Log(loggers, logrus.InfoLevel, fmt.Sprintf("Report generated: %s", reportFilePath))
}

func getDiskMountPoints(queryResults map[string]map[string]model.Value) []string {
	mountPoints := make(map[string]struct{})
	for _, platformResults := range queryResults {
		for _, result := range platformResults {
			matrix, ok := result.(model.Matrix)
			if !ok {
				continue
			}
			for _, sampleStream := range matrix {
				mountPoint := string(sampleStream.Metric["mountpoint"])
				if mountPoint != "" {
					mountPoints[mountPoint] = struct{}{}
				}
			}
		}
	}
	var mounts []string
	for mountPoint := range mountPoints {
		mounts = append(mounts, mountPoint)
	}
	return mounts
}

func calculateDiskStats(platformResults map[string]model.Value, host, mountPoint string) (float64, float64, float64) {
	result, ok := platformResults["DiskUsage"]
	if !ok || result == nil {
		return 0, 0, 0
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return 0, 0, 0
	}

	var max, min, sum float64
	count := 0
	for _, sampleStream := range matrix {
		if string(sampleStream.Metric["instance"]) != host || string(sampleStream.Metric["mountpoint"]) != mountPoint {
			continue
		}
		for _, sample := range sampleStream.Values {
			value := float64(sample.Value)
			if count == 0 || value > max {
				max = value
			}
			if count == 0 || value < min {
				min = value
			}
			sum += value
			count++
		}
	}

	if count == 0 {
		return 0, 0, 0
	}

	avg := sum / float64(count)
	return max, min, avg
}

func writeDiskValues(file *excelize.File, sheetName string, rowIndex, colIndex int, max, min, avg float64, dataStyle int) {
	max *= 100
	min *= 100
	avg *= 100
	format := "%.2f%%"

	for i, value := range []float64{max, min, avg} {
		cell, _ := excelize.CoordinatesToCellName(colIndex+i, rowIndex)
		file.SetCellValue(sheetName, cell, fmt.Sprintf(format, value))
		file.SetCellStyle(sheetName, cell, cell, dataStyle)
	}
}

func extractHostsAndNames(platformResults map[string]model.Value) map[string]string {
	hosts := make(map[string]string)
	for _, result := range platformResults {
		matrix, ok := result.(model.Matrix)
		if !ok {
			continue
		}
		for _, sampleStream := range matrix {
			host := string(sampleStream.Metric["instance"])
			name := string(sampleStream.Metric["name"])
			hosts[host] = name
		}
	}
	return hosts
}

func listMetrics() []string {
	return []string{"CPUUsage", "MemoryUsage", "NetworkUsage"}
}

func calculateStats(platformResults map[string]model.Value, host, metric string) (float64, float64, float64) {
	result, ok := platformResults[metric]
	if !ok || result == nil {
		return 0, 0, 0
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return 0, 0, 0
	}

	var max, min, sum float64
	count := 0
	for _, sampleStream := range matrix {
		if string(sampleStream.Metric["instance"]) != host {
			continue
		}
		for _, sample := range sampleStream.Values {
			value := float64(sample.Value)
			if count == 0 || value > max {
				max = value
			}
			if count == 0 || value < min {
				min = value
			}
			sum += value
			count++
		}
	}

	if count == 0 {
		return 0, 0, 0
	}

	avg := sum / float64(count)
	return max, min, avg
}

func writeMetricValues(file *excelize.File, sheetName string, rowIndex, colIndex int, metric string, max, min, avg float64, dataStyle int) {
	var format string
	if metric == "CPUUsage" || metric == "MemoryUsage" || metric == "DiskUsage" {
		// 将小数转换为百分比
		max *= 100
		min *= 100
		avg *= 100
		format = "%.2f%%"
	} else if metric == "NetworkUsage" {
		// NetworkUsage是MB/s，所以按需要进行转换
		max, min, avg = max/(1024*1024), min/(1024*1024), avg/(1024*1024)
		format = "%.2f"
	} else {
		format = "%.2f"
	}

	// 写入数据并设置样式
	for i, value := range []float64{max, min, avg} {
		cell, _ := excelize.CoordinatesToCellName(colIndex+i, rowIndex)
		file.SetCellValue(sheetName, cell, fmt.Sprintf(format, value))
		file.SetCellStyle(sheetName, cell, cell, dataStyle)
	}
}

func adjustColumnWidths(file *excelize.File, sheetName string, totalCols, totalRows int) {
	for colIndex := 1; colIndex <= totalCols; colIndex++ {
		maxWidth := 0
		for rowIndex := 1; rowIndex <= totalRows; rowIndex++ {
			cell, _ := excelize.CoordinatesToCellName(colIndex, rowIndex)
			value, _ := file.GetCellValue(sheetName, cell)
			if len(value) > maxWidth {
				maxWidth = len(value)
			}
		}
		col, _ := excelize.ColumnNumberToName(colIndex)
		file.SetColWidth(sheetName, col, col, float64(maxWidth)+2)
	}
}
