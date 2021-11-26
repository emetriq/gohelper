package log

import (
	"fmt"
	"strings"

	"net"
	"os"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/rcrowley/go-metrics"
)

var GlobalMetricsPrefix string

//SetGlobalMetricsPrefix set prefix for graphite metrics
//Should be called on application init phase
func SetGlobalMetricsPrefix(name string) {
	GlobalMetricsPrefix = name
}

// GetHostAddr returns the TCPAddr for the given host
func GetHostAddr(host string) *net.TCPAddr {
	if addr, err := net.ResolveTCPAddr("tcp", host); err == nil {
		return addr
	} else {
		panic(err)
	}
}

// GetHostname returns the machine's host name, either as determined by the system, or by using the first IP address of the first non-loopback network interface.
// If the hostname can not be determined, “unkown-ip-address” is returned.
func GetHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	if interfaces, err := net.Interfaces(); err == nil {
		for _, networkInterface := range interfaces {
			if (networkInterface.Flags & net.FlagLoopback) == 0 {
				if addresses, err := networkInterface.Addrs(); err == nil {
					for _, address := range addresses {
						return address.String()
					}
				}
			}
		}
	}
	return "unknown-ip-address"
}

// CreateMetricsRegistry creates a Child Registry with the hostname as prefix
func CreateMetricsRegistry() metrics.Registry {
	hostname := GetHostname()
	hostname = strings.ReplaceAll(hostname, ".", "_")
	return metrics.NewPrefixedChildRegistry(metrics.NewRegistry(), fmt.Sprintf("%s.", hostname))
}

// StartReporter starts the graphite reporter
func StartReporter(graphiteHost string, appName string) {
	addr := GetHostAddr(graphiteHost)
	go metrics.CaptureRuntimeMemStats(metrics.DefaultRegistry, 5*time.Second) // updates the RuntimeMemStats every 5 seconds, the registy does not matter
	go graphite.Graphite(metrics.DefaultRegistry, 10*time.Second, fmt.Sprintf("%s.%s", GlobalMetricsPrefix, appName), addr)
}

// GetCounter creates a new counter or returns the existing counter with the given name from the default registry.
func GetCounter(metricsName string) metrics.Counter {
	return metrics.GetOrRegisterCounter(metricsName, metrics.DefaultRegistry)
}

// GetHistogram creates a new histogram (or returns an existing one with the given name) with an exponential decay sample with default parameters.
func GetHistogram(metricName string) metrics.Histogram {
	return metrics.GetOrRegisterHistogram(metricName, metrics.DefaultRegistry, metrics.NewExpDecaySample(1028, 0.015))
}

// GetMeter creates a new meter or returns the existing meter with the given name from the default registry.
func GetMeter(metricName string) metrics.Meter {
	return metrics.GetOrRegisterMeter(metricName, metrics.DefaultRegistry)
}

// CreateGauge creates a new gauge that reports values returned by the given function.
func CreateGauge(metricName string, value func() int64) {
	gauge := metrics.NewFunctionalGauge(value)
	metrics.Register(metricName, gauge)
}

// CreateGauge creates a new gauge or returns the existing gauge with the given name from the default registry.
func GetGauge(metricName string) metrics.Gauge {
	return metrics.GetOrRegisterGauge(metricName, metrics.DefaultRegistry)
}

// CreateTopicMeterMap creates a map of Meters (topic -> Meter).
func CreateTopicMeterMap(fmtString string, topics []string) map[string]metrics.Meter {
	meterMap := map[string]metrics.Meter{}
	for _, topic := range topics {
		meterName := fmt.Sprintf(fmtString, topic)
		Logger.Infof("created meter: %s", meterName)
		meterMap[topic] = metrics.GetOrRegisterMeter(meterName, metrics.DefaultRegistry)
	}
	return meterMap
}

func MeterMapToSlice(meterMap map[string]metrics.Meter) []metrics.Meter {
	meters := make([]metrics.Meter, 0, len(meterMap))
	for _, m := range meterMap {
		meters = append(meters, m)
	}
	return meters
}
