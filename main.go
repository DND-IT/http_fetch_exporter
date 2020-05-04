package main

import (
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	namespace = "http_fetcher" // For Prometheus metrics.
)

func main() {

	const pidFileHelpText = `Path to HAProxy pid file.`

	var (
		// r              = reader.New("testdata/haproxy.log")
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9102").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		pidFile       = kingpin.Flag("pid-file", pidFileHelpText).Default("").String()
	)

	var url = "https://feed-prod.unitycms.io/image/ocroped/2001,2000,1000,1000,0,0/sjXZsL8l2pM/14wM1Tp84-I9crpvV6Nd9p.jpg"

	promLogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promLogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promLogConfig)

	var err = level.Info(logger).Log("msg", "Starting waf_exporter", "version", version.Info())
	if err != nil {
		log.Fatal(err)
	}
	err = level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())
	if err != nil {
		log.Fatal(err)
	}

	exporter, err := NewExporter(serverMetrics, url, logger)
	if err != nil {
		err = level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("http_fetch_exporter"))

	if *pidFile != "" {
		procExporter := prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
			PidFn: func() (int, error) {
				content, err := ioutil.ReadFile(*pidFile)
				if err != nil {
					return 0, fmt.Errorf("can't read pid file: %s", err)
				}
				value, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					return 0, fmt.Errorf("can't parse pid file: %s", err)
				}
				return value, nil
			},
			Namespace: namespace,
		})
		prometheus.MustRegister(procExporter)
	}

	err = level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var _, err = w.Write([]byte(`<html>
             <head><title>http fetch metrics Exporter</title></head>
             <body>
             <h1>http fetch metrics Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
		if err != nil {
			log.Fatal(err)
		}
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		err = level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

}
