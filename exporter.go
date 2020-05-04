package main

import (
	"github.com/DND-IT/http_fetch_exporter/collector"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type metrics map[int]*prometheus.Desc

var serverLabelNames = []string{"server_ip"}

func newServerMetric(metricName string, docString string, constLabels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "target", metricName), docString, serverLabelNames, constLabels)
}

var (
	serverMetrics = metrics{
		0: newServerMetric("fetch_count", "total count of fetches", nil),
		1: newServerMetric("avg_fetch_time", "avg content fetch time", nil),
		2: newServerMetric("min_fetch_time", "min content fetch time", nil),
		3: newServerMetric("max_fetch_time", "max content fetch time", nil),
	}
)

type Exporter struct {
	mutex         sync.RWMutex
	collector     *collector.Engine
	tenantMetrics map[int]*prometheus.Desc
	logger        log.Logger
}

// Describe describes all the metrics ever exported by the HAProxy exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range serverMetrics {
		ch <- m
	}
}

// Collect fetches the stats from configured HAProxy location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	e.Scrape(ch)

}

func (e *Exporter) Scrape(ch chan<- prometheus.Metric) {
	var server = e.collector.DumpServer()
	for _, v := range server {
		ch <- prometheus.MustNewConstMetric(serverMetrics[0], prometheus.CounterValue, float64(v.FetchCount), v.IP)
		ch <- prometheus.MustNewConstMetric(serverMetrics[1], prometheus.GaugeValue, v.AVGFetchTime, v.IP)
		ch <- prometheus.MustNewConstMetric(serverMetrics[2], prometheus.GaugeValue, v.MinFetchTime, v.IP)
		ch <- prometheus.MustNewConstMetric(serverMetrics[3], prometheus.GaugeValue, v.MaxFetchTime, v.IP)
	}

}

// NewExporter returns an initialized Exporter.
func NewExporter(selectedServerMetrics map[int]*prometheus.Desc, url string, logger log.Logger) (*Exporter, error) {

	var err error

	if err != nil {
		return nil, err
	}

	// start collector
	var c = collector.New(url)
	c.Start()

	return &Exporter{
		tenantMetrics: selectedServerMetrics,
		logger:        logger,
		collector:     c,
	}, nil
}
