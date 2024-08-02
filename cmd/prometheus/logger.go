package prometheusdb

import (
	"time"

	sniffer "github.com/kelcheone/netrics/cmd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	bytesMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "data_usage_bytes",
			Help: "Number of bytes transferred",
		}, []string{"ip", "direction"},
	)
	lastUpdateMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "data_usage_last_update",
			Help: "Timestamp of the last update",
		}, []string{"ip"},
	)
)

func LogUsage(usage sniffer.Usage, ip string, bytes int64, incoming bool, timestamp time.Time) {
	dataUsage, exists := usage[ip]
	if !exists {
		dataUsage = &sniffer.DataUsage{Timestamp: timestamp}
		usage[ip] = dataUsage
	}

	dataUsage.TotalBytes += bytes

	if incoming {
		usage[ip].IncomingBytes += bytes
	} else {
		usage[ip].OutGoingBytes += bytes
	}
	dataUsage.Timestamp = timestamp

	direction := "outgoing"

	if incoming {
		direction = "incoming"
	}

	bytesMetric.With(prometheus.Labels{
		"ip":        ip,
		"direction": direction,
	}).Add(float64(bytes))

	lastUpdateMetric.With(prometheus.Labels{"ip": ip}).Set(float64(timestamp.Unix()))
}
