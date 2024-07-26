package query

import (
	"fmt"
)

// GetQueries returns a map of Prometheus queries with a specified window size.
func GetQueries(window string) map[string]string {
	return map[string]string{
		"CPUUsage":     fmt.Sprintf("1 - avg by (instance,name) (irate(node_cpu_seconds_total{mode='idle'}[%s]))", window),
		"MemoryUsage":  "1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)",
		"DiskUsage":    "1 - (node_filesystem_free_bytes{fstype=~\"ext[.]?|xfs\",mountpoint=~\"^/$|^/data$\", mountpoint!=\"\"} / node_filesystem_size_bytes{fstype=~\"ext[.]?|xfs\",mountpoint=~\"^/$|^/data$\", mountpoint!=\"\"})",
		"NetworkUsage": fmt.Sprintf("sum(rate(node_network_receive_bytes_total[%s]) + rate(node_network_transmit_bytes_total[%s])) by (instance,name)", window, window),
	}
}
