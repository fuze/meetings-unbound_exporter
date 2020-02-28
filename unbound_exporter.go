package main

import (
  "bytes"
  "errors"
  "flag"
  "fmt"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
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
	metrics map[string]float64

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

	timeUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "time_up"),
		"Number of seconds process is running",
		nil, nil,
	)

	memTotalSbrk = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mem_total_sbrk"),
		"Amount of sbrk memory in bytes",
		nil, nil,
	)

	memCacheRrset = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mem_cache_rrset"),
		"Amount of cache rrset memory in bytes",
		nil, nil,
	)

	memCacheMessage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mem_cache_message"),
		"Amount of cache message memory in bytes",
		nil, nil,
	)

	memModIterator = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mem_mod_iterator"),
		"Amount of memory allocated to the iterator module in bytes",
		nil, nil,
	)

	memModValidator = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mem_mod_validator"),
		"Amount of memory allocated to the validator module in bytes",
		nil, nil,
	)

	hist0usTo1us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_0us_to_1us"),
		"Number of requests answered in 0us to 1us",
		nil, nil,
	)

	hist1usTo2us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_1us_to_2us"),
		"Number of requests answered in 1us to 2us",
		nil, nil,
	)

	hist2usTo4us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_2us_to_4us"),
		"Number of requests answered in 2us to 4us",
		nil, nil,
	)

	hist4usTo8us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_4us_to_8us"),
		"Number of requests answered in 2us to 4us",
		nil, nil,
	)

	hist8usTo16us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_8us_to_16us"),
		"Number of requests answered in 8us to 16us",
		nil, nil,
	)

	hist16usTo32us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_16us_to_32us"),
		"Number of requests answered in 16us to 32us",
		nil, nil,
	)

	hist32usTo64us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_32us_to_64us"),
		"Number of requests answered in 32us to 64us",
		nil, nil,
	)

	hist64usTo128us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_64us_to_128us"),
		"Number of requests answered in 64us to 128us",
		nil, nil,
	)

	hist128usTo256us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_128us_to_256us"),
		"Number of requests answered in 128us to 256us",
		nil, nil,
	)

	hist256usTo512us = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_256us_to_512us"),
		"Number of requests answered in 256us to 512us",
		nil, nil,
	)

	hist512usTo1ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_512us_to_1ms"),
		"Number of requests answered in 512us to 1ms",
		nil, nil,
	)

	hist1msTo2ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_1ms_to_2ms"),
		"Number of requests answered in 1ms to 2ms",
		nil, nil,
	)

	hist2msTo4ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_2ms_to_4ms"),
		"Number of requests answered in 2ms to 4ms",
		nil, nil,
	)

	hist4msTo8ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_4ms_to_8ms"),
		"Number of requests answered in 4ms to 8ms",
		nil, nil,
	)

	hist8msTo16ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_8ms_to_16ms"),
		"Number of requests answered in 8ms to 16ms",
		nil, nil,
	)

	hist16msTo32ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_16ms_to_32ms"),
		"Number of requests answered in 16ms to 32ms",
		nil, nil,
	)

	hist32msTo64ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_32ms_to_64ms"),
		"Number of requests answered in 32ms to 64ms",
		nil, nil,
	)

	hist64msTo128ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_64ms_to_128ms"),
		"Number of requests answered in 64ms to 128ms",
		nil, nil,
	)

	hist128msTo256ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_128ms_to_256ms"),
		"Number of requests answered in 128ms to 256ms",
		nil, nil,
	)

	hist256msTo512ms = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_256ms_to_512ms"),
		"Number of requests answered in 256ms to 512ms",
		nil, nil,
	)

	hist512msTo1s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_512ms_to_1s"),
		"Number of requests answered in 512ms to 1s",
		nil, nil,
	)

	hist1sTo2s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_1s_to_2s"),
		"Number of requests answered in 1s to 2s",
		nil, nil,
	)

	hist2sTo4s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_2s_to_4s"),
		"Number of requests answered in 2s to 4s",
		nil, nil,
	)

	hist4sTo8s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_4s_to_8s"),
		"Number of requests answered in 4s to 8s",
		nil, nil,
	)

	hist8sTo16s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_8s_to_16s"),
		"Number of requests answered in 8s to 16s",
		nil, nil,
	)

	hist16sTo32s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_16s_to_32s"),
		"Number of requests answered in 16s to 32s",
		nil, nil,
	)

	hist32sTo64s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_32s_to_64s"),
		"Number of requests answered in 32s to 64s",
		nil, nil,
	)

	hist64sTo128s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_64s_to_128s"),
		"Number of requests answered in 64s to 128s",
		nil, nil,
	)

	hist128sTo256s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_128s_to_256s"),
		"Number of requests answered in 128s to 256s",
		nil, nil,
	)

	hist256sTo512s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_256s_to_512s"),
		"Number of requests answered in 256s to 512s",
		nil, nil,
	)

	hist512sTo1024s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_512s_to_1024s"),
		"Number of requests answered in 512s to 1024s",
		nil, nil,
	)

	hist1024sTo2048s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_1024s_to_2048s"),
		"Number of requests answered in 1024s to 2048s",
		nil, nil,
	)

	hist2048sTo4096s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_2048s_to_4096s"),
		"Number of requests answered in 2048s to 4096s",
		nil, nil,
	)

	hist4096sTo8192s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_4096s_to_8192s"),
		"Number of requests answered in 4096s to 8192s",
		nil, nil,
	)

	hist8192sTo16384s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_8192s_to_16384s"),
		"Number of requests answered in 8192s to 16384s",
		nil, nil,
	)

	hist16384sTo32768s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_16384s_to_32768s"),
		"Number of requests answered in 16384s to 32768s",
		nil, nil,
	)

	hist32768sTo65536s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_32768s_to_65536s"),
		"Number of requests answered in 32768s to 65536s",
		nil, nil,
	)

	hist65536sTo131072s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_65536s_to_131072s"),
		"Number of requests answered in 65536s to 131072s",
		nil, nil,
	)

	hist131072sTo262144s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_131072s_to_262144s"),
		"Number of requests answered in 131072s to 262144s",
		nil, nil,
	)

	hist262144sTo524288s = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "histogram_262144s_to_524288s"),
		"Number of requests answered in 262144s to 524288s",
		nil, nil,
	)

	numQueryTypeA = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_a"),
		"Number of requests for A records",
		nil, nil,
	)

	numQueryTypeAAAA = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_aaaa"),
		"Number of requests for AAAA records",
		nil, nil,
	)

	numQueryTypePTR = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_ptr"),
		"Number of requests for PTR records",
		nil, nil,
	)

	numQueryTypeSRV = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_srv"),
		"Number of requests for SRV records",
		nil, nil,
	)

	numQueryTypeMX = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_mx"),
		"Number of requests for MX records",
		nil, nil,
	)

	numQueryTypeNS = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_type_ns"),
		"Number of requests for NS records",
		nil, nil,
	)

	numQueryClassIN = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_class_in"),
		"Number of queries in the IN class",
		nil, nil,
	)

	numQueryOpcodeQUERY = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_opcode_query"),
		"Number of queries containing the QUERY opcode",
		nil, nil,
	)

	numQueryTcp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_tcp"),
		"Number of TCP queries",
		nil, nil,
	)

	numQueryTcpOut = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_tcp_out"),
		"Number of outgoing TCP queries",
		nil, nil,
	)

	numQueryIPv6 = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_ipv6"),
		"Number of IPv6 queries",
		nil, nil,
	)

	numQueryFlagsQR = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_qr"),
		"Number of queries with the QR flag",
		nil, nil,
	)

	numQueryFlagsAA = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_aa"),
		"Number of queries with the AA flag",
		nil, nil,
	)

	numQueryFlagsTC = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_tc"),
		"Number of queries with the TC flag",
		nil, nil,
	)

	numQueryFlagsRD = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_rd"),
		"Number of queries with the RD flag",
		nil, nil,
	)

	numQueryFlagsRA = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_ra"),
		"Number of queries with the RA flag",
		nil, nil,
	)

	numQueryFlagsZ = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_z"),
		"Number of queries with the Z flag",
		nil, nil,
	)

	numQueryFlagsAD = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_ad"),
		"Number of queries with the AD flag",
		nil, nil,
	)

	numQueryFlagsCD = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_flags_cd"),
		"Number of queries with the CD flag",
		nil, nil,
	)

	numQueryEdnsPresent = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_edns_present"),
		"Number of EDNS queries",
		nil, nil,
	)

	numQueryEdnsDO = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_query_edns_do"),
		"Number of edns queries with the DO flag",
		nil, nil,
	)

	numAnswerRcodeNOERROR = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_noerror"),
		"Number of answers with rcode NOERROR",
		nil, nil,
	)

	numAnswerRcodeFORMERR = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_formerr"),
		"Number of answers with rcode FORMERR",
		nil, nil,
	)

	numAnswerRcodeSERVFAIL = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_servfail"),
		"Number of answers with rcode SERVFAIL",
		nil, nil,
	)

	numAnswerRcodeNXDOMAIN = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_nxdomain"),
		"Number of answers with rcode NXDOMAIN",
		nil, nil,
	)

	numAnswerRcodeNOTIMPL = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_notimpl"),
		"Number of answers with rcode NOTIMPL",
		nil, nil,
	)

	numAnswerRcodeREFUSED = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_refused"),
		"Number of answers with rcode REFUSED",
		nil, nil,
	)

	numAnswerRcodeNoData = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_rcode_nodata"),
		"Number of answers with rcode nodata",
		nil, nil,
	)

	numAnswerSecure = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_secure"),
		"Number of secure answers",
		nil, nil,
	)

	numAnswerBogus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_answer_bogus"),
		"Number of bogus answers",
		nil, nil,
	)

	numRrsetBogus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "num_rrset_bogus"),
		"Number of bogus rrsets",
		nil, nil,
	)

	unwantedQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "unwanted_queries"),
		"Number of unwanted queries",
		nil, nil,
	)

	unwantedReplies = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "unwanted_replies"),
		"Number of unwanted replies",
		nil, nil,
	)

	msgCacheCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "msg_cache_count"),
		"Number of cached messages",
		nil, nil,
	)

	rrsetCacheCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "rrset_cache_count"),
		"Number of cached rrsets",
		nil, nil,
	)

	infraCacheCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "infra_cache_count"),
		"Number of cached infra items",
		nil, nil,
	)

	keyCacheCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "key_cache_count"),
		"Number of cached keys",
		nil, nil,
	)
)

type Exporter struct {
	Command string
}

func lookupMetric(key string) float64 {
	if value, exists := metrics[key]; exists {
		return value
	} else {
		return 0
	}
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
	ch <- timeUp
	ch <- memTotalSbrk
	ch <- memCacheRrset
	ch <- memCacheMessage
	ch <- memModIterator
	ch <- memModValidator
	ch <- hist0usTo1us
	ch <- hist1usTo2us
	ch <- hist2usTo4us
	ch <- hist4usTo8us
	ch <- hist8usTo16us
	ch <- hist16usTo32us
	ch <- hist32usTo64us
	ch <- hist64usTo128us
	ch <- hist128usTo256us
	ch <- hist256usTo512us
	ch <- hist512usTo1ms
	ch <- hist1msTo2ms
	ch <- hist2msTo4ms
	ch <- hist4msTo8ms
	ch <- hist8msTo16ms
	ch <- hist16msTo32ms
	ch <- hist32msTo64ms
	ch <- hist64msTo128ms
	ch <- hist128msTo256ms
	ch <- hist256msTo512ms
	ch <- hist512msTo1s
	ch <- hist1sTo2s
	ch <- hist2sTo4s
	ch <- hist4sTo8s
	ch <- hist8sTo16s
	ch <- hist16sTo32s
	ch <- hist32sTo64s
	ch <- hist64sTo128s
	ch <- hist128sTo256s
	ch <- hist256sTo512s
	ch <- hist512sTo1024s
	ch <- hist1024sTo2048s
	ch <- hist2048sTo4096s
	ch <- hist4096sTo8192s
	ch <- hist8192sTo16384s
	ch <- hist16384sTo32768s
	ch <- hist32768sTo65536s
	ch <- hist65536sTo131072s
	ch <- hist131072sTo262144s
	ch <- hist262144sTo524288s
	ch <- numQueryTypeA
	ch <- numQueryTypeAAAA
	ch <- numQueryTypePTR
	ch <- numQueryTypeSRV
	ch <- numQueryTypeMX
	ch <- numQueryTypeNS
	ch <- numQueryClassIN
	ch <- numQueryOpcodeQUERY
	ch <- numQueryTcp
	ch <- numQueryTcpOut
	ch <- numQueryIPv6
	ch <- numQueryFlagsQR
	ch <- numQueryFlagsAA
	ch <- numQueryFlagsTC
	ch <- numQueryFlagsRD
	ch <- numQueryFlagsRA
	ch <- numQueryFlagsZ
	ch <- numQueryFlagsAD
	ch <- numQueryFlagsCD
	ch <- numQueryEdnsPresent
	ch <- numQueryEdnsDO
	ch <- numAnswerRcodeNOERROR
	ch <- numAnswerRcodeFORMERR
	ch <- numAnswerRcodeSERVFAIL
	ch <- numAnswerRcodeNXDOMAIN
	ch <- numAnswerRcodeNOTIMPL
	ch <- numAnswerRcodeREFUSED
	ch <- numAnswerRcodeNoData
	ch <- numAnswerSecure
	ch <- numAnswerBogus
	ch <- numRrsetBogus
	ch <- unwantedQueries
	ch <- unwantedReplies
	ch <- msgCacheCount
	ch <- rrsetCacheCount
	ch <- infraCacheCount
	ch <- keyCacheCount
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var (
		stdout_buf bytes.Buffer
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
			value = 0
		}
		metrics[tokens[0]] = value
	}

	ch <- prometheus.MustNewConstMetric(
		totalNumQueries, prometheus.CounterValue,
		lookupMetric("total.num.queries"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumCacheHits, prometheus.CounterValue,
		lookupMetric("total.num.cachehits"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumCacheMiss, prometheus.CounterValue,
		lookupMetric("total.num.cachemiss"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumPrefetch, prometheus.CounterValue,
		lookupMetric("total.num.prefetch"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalNumRecursiveReplies, prometheus.CounterValue,
		lookupMetric("total.num.recursivereplies"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistAvg, prometheus.GaugeValue,
		lookupMetric("total.requestlist.avg"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistMax, prometheus.GaugeValue,
		lookupMetric("total.requestlist.max"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistOverwritten, prometheus.CounterValue,
		lookupMetric("total.requestlist.overwritten"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistExceeded, prometheus.CounterValue,
		lookupMetric("total.requestlist.exceeded"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistCurrentAll, prometheus.GaugeValue,
		lookupMetric("total.requestlist.current.all"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRequestlistCurrentUser, prometheus.GaugeValue,
		lookupMetric("total.requestlist.current.user"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRecursionTimeAvg, prometheus.GaugeValue,
		lookupMetric("total.recursion.time.avg"),
	)

	ch <- prometheus.MustNewConstMetric(
		totalRecurseTimeMedian, prometheus.GaugeValue,
		lookupMetric("total.recursion.time.median"),
	)
	ch <- prometheus.MustNewConstMetric(
		timeUp, prometheus.CounterValue,
		lookupMetric("time.up"),
	)

	ch <- prometheus.MustNewConstMetric(
		memTotalSbrk, prometheus.GaugeValue,
		lookupMetric("mem.total.sbrk"),
	)

	ch <- prometheus.MustNewConstMetric(
		memCacheRrset, prometheus.GaugeValue,
		lookupMetric("mem.cache.rrset"),
	)

	ch <- prometheus.MustNewConstMetric(
		memCacheMessage, prometheus.GaugeValue,
		lookupMetric("mem.cache.message"),
	)

	ch <- prometheus.MustNewConstMetric(
		memModIterator, prometheus.GaugeValue,
		lookupMetric("mem.mod.iterator"),
	)

	ch <- prometheus.MustNewConstMetric(
		memModValidator, prometheus.GaugeValue,
		lookupMetric("mem.mod.validator"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist0usTo1us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000000.to.000000.000001"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist1usTo2us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000001.to.000000.000002"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist2usTo4us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000002.to.000000.000004"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist4usTo8us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000004.to.000000.000008"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist8usTo16us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000008.to.000000.000016"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist16usTo32us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000016.to.000000.000032"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist32usTo64us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000032.to.000000.000064"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist64usTo128us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000064.to.000000.000128"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist128usTo256us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000128.to.000000.000256"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist256usTo512us, prometheus.CounterValue,
		lookupMetric("histogram.000000.000256.to.000000.000512"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist512usTo1ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.000512.to.000000.001024"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist1msTo2ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.001024.to.000000.002048"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist2msTo4ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.002048.to.000000.004096"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist4msTo8ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.004096.to.000000.008192"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist8msTo16ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.008192.to.000000.016384"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist16msTo32ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.016384.to.000000.032768"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist32msTo64ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.032768.to.000000.065536"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist64msTo128ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.065536.to.000000.131072"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist128msTo256ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.131072.to.000000.262144"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist256msTo512ms, prometheus.CounterValue,
		lookupMetric("histogram.000000.262144.to.000000.524288"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist512msTo1s, prometheus.CounterValue,
		lookupMetric("histogram.000000.524288.to.000001.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist1sTo2s, prometheus.CounterValue,
		lookupMetric("histogram.000001.000000.to.000002.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist2sTo4s, prometheus.CounterValue,
		lookupMetric("histogram.000002.000000.to.000004.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist4sTo8s, prometheus.CounterValue,
		lookupMetric("histogram.000004.000000.to.000008.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist8sTo16s, prometheus.CounterValue,
		lookupMetric("histogram.000008.000000.to.000016.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist16sTo32s, prometheus.CounterValue,
		lookupMetric("histogram.000016.000000.to.000032.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist32sTo64s, prometheus.CounterValue,
		lookupMetric("histogram.000032.000000.to.000064.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist64sTo128s, prometheus.CounterValue,
		lookupMetric("histogram.000064.000000.to.000128.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist128sTo256s, prometheus.CounterValue,
		lookupMetric("histogram.000128.000000.to.000256.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist256sTo512s, prometheus.CounterValue,
		lookupMetric("histogram.000256.000000.to.000512.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist512sTo1024s, prometheus.CounterValue,
		lookupMetric("histogram.000512.000000.to.001024.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist1024sTo2048s, prometheus.CounterValue,
		lookupMetric("histogram.001024.000000.to.002048.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist2048sTo4096s, prometheus.CounterValue,
		lookupMetric("histogram.002048.000000.to.004096.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist4096sTo8192s, prometheus.CounterValue,
		lookupMetric("histogram.004096.000000.to.008192.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist8192sTo16384s, prometheus.CounterValue,
		lookupMetric("histogram.008192.000000.to.016384.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist16384sTo32768s, prometheus.CounterValue,
		lookupMetric("histogram.016384.000000.to.032768.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist32768sTo65536s, prometheus.CounterValue,
		lookupMetric("histogram.032768.000000.to.065536.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist65536sTo131072s, prometheus.CounterValue,
		lookupMetric("histogram.065536.000000.to.131072.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist131072sTo262144s, prometheus.CounterValue,
		lookupMetric("histogram.131072.000000.to.262144.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		hist262144sTo524288s, prometheus.CounterValue,
		lookupMetric("histogram.262144.000000.to.524288.000000"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypeA, prometheus.CounterValue,
		lookupMetric("num.query.type.A"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypeAAAA, prometheus.CounterValue,
		lookupMetric("num.query.type.AAAA"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypePTR, prometheus.CounterValue,
		lookupMetric("num.query.type.PTR"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypeSRV, prometheus.CounterValue,
		lookupMetric("num.query.type.SRV"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypeMX, prometheus.CounterValue,
		lookupMetric("num.query.type.MX"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTypeNS, prometheus.CounterValue,
		lookupMetric("num.query.type.NS"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryClassIN, prometheus.CounterValue,
		lookupMetric("num.query.class.IN"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryOpcodeQUERY, prometheus.CounterValue,
		lookupMetric("num.query.opcode.QUERY"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTcp, prometheus.CounterValue,
		lookupMetric("num.query.tcp"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryTcpOut, prometheus.CounterValue,
		lookupMetric("num.query.tcpout"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryIPv6, prometheus.CounterValue,
		lookupMetric("num.query.ipv6"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsQR, prometheus.CounterValue,
		lookupMetric("num.query.flags.QR"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsAA, prometheus.CounterValue,
		lookupMetric("num.query.flags.AA"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsTC, prometheus.CounterValue,
		lookupMetric("num.query.flags.TC"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsRD, prometheus.CounterValue,
		lookupMetric("num.query.flags.RD"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsRA, prometheus.CounterValue,
		lookupMetric("num.query.flags.RA"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsZ, prometheus.CounterValue,
		lookupMetric("num.query.flags.Z"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsAD, prometheus.CounterValue,
		lookupMetric("num.query.flags.AD"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryFlagsCD, prometheus.CounterValue,
		lookupMetric("num.query.flags.CD"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryEdnsPresent, prometheus.CounterValue,
		lookupMetric("num.query.edns.present"),
	)

	ch <- prometheus.MustNewConstMetric(
		numQueryEdnsDO, prometheus.CounterValue,
		lookupMetric("num.query.edns.DO"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeNOERROR, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.NOERROR"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeFORMERR, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.FORMERR"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeSERVFAIL, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.SERVFAIL"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeNXDOMAIN, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.NXDOMAIN"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeNOTIMPL, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.NOTIMPL"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeREFUSED, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.REFUSED"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerRcodeNoData, prometheus.CounterValue,
		lookupMetric("num.answer.rcode.nodata"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerSecure, prometheus.CounterValue,
		lookupMetric("num.answer.secure"),
	)

	ch <- prometheus.MustNewConstMetric(
		numAnswerBogus, prometheus.CounterValue,
		lookupMetric("num.answer.bogus"),
	)

	ch <- prometheus.MustNewConstMetric(
		numRrsetBogus, prometheus.CounterValue,
		lookupMetric("num.rrset.bogus"),
	)

	ch <- prometheus.MustNewConstMetric(
		unwantedQueries, prometheus.CounterValue,
		lookupMetric("unwanted.queries"),
	)

	ch <- prometheus.MustNewConstMetric(
		unwantedReplies, prometheus.CounterValue,
		lookupMetric("unwanted.replies"),
	)

	ch <- prometheus.MustNewConstMetric(
		msgCacheCount, prometheus.GaugeValue,
		lookupMetric("msg.cache.count"),
	)

	ch <- prometheus.MustNewConstMetric(
		rrsetCacheCount, prometheus.GaugeValue,
		lookupMetric("rrset.cache.count"),
	)

	ch <- prometheus.MustNewConstMetric(
		infraCacheCount, prometheus.GaugeValue,
		lookupMetric("infra.cache.count"),
	)

	ch <- prometheus.MustNewConstMetric(
		keyCacheCount, prometheus.GaugeValue,
		lookupMetric("key.cache.count"),
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

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Unbound Exporter</title></head>
			<body>
			<h1>Unbound exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
