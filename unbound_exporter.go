package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	namespace = "unbound"
)

var (
	totalNumQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_num_queries"),
		"Number of queries for all threads.",
		nil, nil,
	)

	totalNumCacheHits = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_num_cache_hits"),
		"Number of cache hits for all threads.",
		nil, nil,
	)

	totalNumCacheMiss = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_num_cache_miss"),
		"Number of cache misses for all threads.",
		nil, nil,
	)

	totalNumPrefetch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_num_prefetch"),
		"Number of prefetches for all threads.",
		nil, nil,
	)

	totalNumRecursiveReplies = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_num_recursive_replies"),
		"Number of recursive replies for all threads.",
		nil, nil,
	)

	totalRequestlistAvg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_avg"),
		"Average requestlist size for all threads.",
		nil, nil,
	)

	totalRequestlistMax = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_max"),
		"Maximum requestlist size for all threads",
		nil, nil,
	)

	totalRequestlistOverwritten = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_overwritten"),
		"Number of items overwritten in requestlist for all threads",
		nil, nil,
	)

	totalRequestlistExceeded = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_exceeded"),
		"Number of items that exceeded the requestlist for all threads",
		nil, nil,
	)

	totalRequestlistCurrentAll = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_current_all"),
		"All current items on the requestlist for all threads",
		nil, nil,
	)

	totalRequestlistCurrentUser = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_requestlist_current_user"),
		"User current items on the requestlist for all threads",
		nil, nil,
	)

	totalRecursionTimeAvg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_recurse_time_avg"),
		"Average time spent recursing",
		nil, nil,
	)

	totalRecurseTimeMedian = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "total_recurse_time_median"),
		"Median time spent recursing",
		nil, nil,
	)
)

type Exporter struct {
	Command string
}

func NewExporter(command string) (*Exporter, error) {
	if _, err := os.Stat(command); err != nil {
		return nil, errors.New("Failed to instantiate exporter: " + err.Error())
	}

	return &Exporter{
		Command: command,
	}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalNumQueries
	ch <- totalNumCacheHits
	ch <- totalNumCacheHits
	ch <- totalNumCacheMiss
	ch <- totalNumPrefetch
	ch <- totalNumRecursiveReplies
	ch <- totalRequestlistAvg
	ch <- totalRequestlistMax
	ch <- totalRequestlistOverwritten
	ch <- totalRequestlistExceeded
	ch <- totalRequestlistCurrentAll
	ch <- totalRequestlistCurrentUser
	ch <- totalRecursionTimeAvg
	ch <- totalRecurseTimeMedian
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var (
		stdout_buf bytes.Buffer
		metrics    map[string]float64
	)

	metrics = make(map[string]float64)

	cmd := exec.Command(e.Command, "stats_noreset")
	cmd.Stdout = &stdout_buf
	if err := cmd.Run(); err != nil {
		log.Fatalln("Failed to collect metrics: " + err.Error())
	}

	for _, metric := range strings.Split(stdout_buf.String(), "\n") {
		if len(metric) == 0 {
			continue
		}
		tokens := strings.Split(metric, "=")
		value, err := strconv.ParseFloat(tokens[1], 8)
		if err != nil {
			log.Warnln("Failed to parse float: " + err.Error())
		}
		metrics[tokens[0]] = value
	}

	ch <- prometheus.MustNewConstMetric(
		totalNumQueries, prometheus.CounterValue, metrics["total.num.queries"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumCacheHits, prometheus.CounterValue, metrics["total.num.cachehits"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumCacheMiss, prometheus.CounterValue, metrics["total.num.cachemiss"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumPrefetch, prometheus.CounterValue, metrics["total.num.prefetch"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumRecursiveReplies, prometheus.CounterValue, metrics["total.num.recursivereplies"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistAvg, prometheus.GaugeValue, metrics["total.requestlist.avg"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistMax, prometheus.GaugeValue, metrics["total.requestlist.max"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistOverwritten, prometheus.CounterValue, metrics["total.requestlist.overwritten"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistExceeded, prometheus.CounterValue, metrics["total.requestlist.exceeded"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistCurrentAll, prometheus.GaugeValue, metrics["total.requestlist.current.all"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistCurrentUser, prometheus.GaugeValue, metrics["total.requestlist.current.user"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRecursionTimeAvg, prometheus.GaugeValue, metrics["total.recursion.time.avg"],
	)

	ch <- prometheus.MustNewConstMetric(
		totalRecurseTimeMedian, prometheus.GaugeValue, metrics["total.recursion.time.median"],
	)

}

func init() {
	prometheus.MustRegister(version.NewCollector("unbound_exporter"))
}

func main() {
	var (
		showVersion   = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", ":9107",
			"Address to listen on for web interface and telemetry.")
		metricsPath = flag.String("web.telemetry-path", "/metrics",
			"Path under which to expose metrics.")
		commandPath = flag.String("unbound.control", "/usr/sbin/unbound-control",
			"Path to unbound-control.")
	)

	flag.Parse()

	if *showVersion {
		fmt.Println(os.Stdout, version.Print("unbound_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting unbound_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter(*commandPath)
	if err != nil {
		log.Fatalln(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Unbound Exporter</title></head>
			<body>
			<h1>Unbound exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening ln", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
