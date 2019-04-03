package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dreampuf/kafka-manager-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	kurls    = flag.String("urls", "http://127.0.0.1:9000", "Kafka Manager URLs. Separate with comma")
	interval = flag.Duration("interval", time.Second*15, "interval of collecting")
	addr     = flag.String("addr", "127.0.0.1:9001", "http server listening address")
)

var (
	rpcDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "rpc_durations_seconds",
			Help: "RPC latency distributions.",
		},
		[]string{},
	)
	topicOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_offset",
			Help: "Topic Partition Offset",
		},
		[]string{"cluster", "topic", "partition"},
	)
	topicSummedOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_summed_offset",
			Help: "Topic Summed Offset",
		},
		[]string{"cluster", "topic"},
	)
	topicChangeRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_change_rate",
			Help: "Topic Change Rate",
		},
		[]string{"cluster", "topic"},
	)
	topicPerPartitionChangeRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_partition_change_rate",
			Help: "Topic Per Partition Change Rate",
		},
		[]string{"cluster", "topic", "partition"},
	)
	consumerTopicLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_lag",
			Help: "Consumer Lag Per Topic",
		},
		[]string{"cluster", "consumer", "topic"},
	)
	consumerTopicPercentageCovered = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_percentage_covered",
			Help: "Consumer Percentage Covered Per Topic",
		},
		[]string{"cluster", "consumer", "topic"},
	)
	consumerTopicPartitionOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_partition_offset",
			Help: "Consumer Topic Offset Per Partition",
		},
		[]string{"cluster", "consumer", "topic", "partition"},
	)
	consumerTopicPartitionLatestOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_partition_latest_offset",
			Help: "Consumer Topic Latest Offset Per Partition",
		},
		[]string{"cluster", "consumer", "topic", "partition"},
	)

	brokerBytesIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "broker_bytes_in_second",
			Help: "Traffic In Per Broker",
		}, []string{"cluster", "host"},
	)
	brokerBytesOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "broker_bytes_out_second",
			Help: "Traffic Out Per Broker",
		}, []string{"cluster", "host"},
	)
	clusterMessageIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_message_in_second",
			Help: "Message In Per Cluster",
		}, []string{"cluster"},
	)
	clusterBytesIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_bytes_in_second",
			Help: "Bytes In Per Cluster",
		}, []string{"cluster"},
	)
	clusterBytesOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_bytes_out_second",
			Help: "Bytes Out Per Cluster",
		}, []string{"cluster"},
	)
	clusterBytesRejected = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_bytes_rejected_second",
			Help: "Bytes Rejected Per Cluster",
		}, []string{"cluster"},
	)
	clusterFailedFetchRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_failed_fetch_request_second",
			Help: "Failed Fetch Request Per Cluster",
		}, []string{"cluster"},
	)
	clusterFailedProduceRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_failed_produce_request_second",
			Help: "Failed Produce Request Per Cluster",
		}, []string{"cluster"},
	)

	topicMessageIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_message_in_second",
			Help: "Message In Per Topic",
		}, []string{"topic"},
	)
	topicBytesIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_bytes_in_second",
			Help: "Bytes In Per Topic",
		}, []string{"topic"},
	)
	topicBytesOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_bytes_out_second",
			Help: "Bytes Out Per Cluster",
		}, []string{"topic"},
	)
	topicBytesRejected = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_bytes_rejected_second",
			Help: "Bytes Rejected Per Cluster",
		}, []string{"topic"},
	)
	topicFailedFetchRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_failed_fetch_request_second",
			Help: "Failed Fetch Request Per Cluster",
		}, []string{"topic"},
	)
	topicFailedProduceRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_failed_produce_request_second",
			Help: "Failed Produce Request Per Cluster",
		}, []string{"topic"},
	)
)

var (
	labelToClusterMetrics = map[string]*prometheus.GaugeVec{
		"Messages in /sec": clusterMessageIn,
		"Bytes in /sec": clusterBytesIn,
		"Bytes out /sec": clusterBytesOut,
		"Bytes rejected /sec": clusterBytesRejected,
		"Failed fetch request /sec": clusterFailedFetchRequest,
		"Failed produce request /sec": clusterFailedProduceRequest,
	}
	labelToTopicMetrics = map[string]*prometheus.GaugeVec{
		"Messages in /sec": topicMessageIn,
		"Bytes in /sec": topicBytesIn,
		"Bytes out /sec": topicBytesOut,
		"Bytes rejected /sec": topicBytesRejected,
		"Failed fetch request /sec": topicFailedFetchRequest,
		"Failed produce request /sec": topicFailedProduceRequest,
	}
)

func init() {
	prometheus.MustRegister(rpcDurations)
	prometheus.MustRegister(topicOffset)
	prometheus.MustRegister(topicSummedOffset)
	prometheus.MustRegister(topicChangeRate)
	prometheus.MustRegister(topicPerPartitionChangeRate)
	prometheus.MustRegister(consumerTopicLag)
	prometheus.MustRegister(consumerTopicPercentageCovered)
	prometheus.MustRegister(consumerTopicPartitionOffset)
	prometheus.MustRegister(consumerTopicPartitionLatestOffset)
	prometheus.MustRegister(brokerBytesIn, brokerBytesOut,
		clusterMessageIn, clusterBytesIn, clusterBytesOut, clusterBytesRejected,
		clusterFailedFetchRequest, clusterFailedProduceRequest)
	prometheus.MustRegister(topicMessageIn,
		topicBytesIn, topicBytesOut, topicBytesRejected,
		topicFailedFetchRequest, topicFailedProduceRequest)
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.TODO())

	// signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		select {
		case <-signalCh:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	g, gctx := errgroup.WithContext(ctx)
	run(g, gctx)

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func run(g *errgroup.Group, ctx context.Context) {
	t := time.NewTicker(*interval)
	g.Go(func() error {
		for {
			select {
			case <-t.C:
				client := &http.Client{}
				wg := &sync.WaitGroup{}
				for _, kmurl := range strings.Split(*kurls, ",") {
					wg.Add(1)
					go collect(wg, ctx, client, kmurl)
				}
				wg.Wait()
			case <-ctx.Done():
				t.Stop()
				return nil
			}
		}
	})

	var srv http.Server
	srv.Addr = *addr
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<a href="/metrics">Metrics</a>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		select {
		case <-ctx.Done():
			srv.Shutdown(ctx)
			log.Print("Server Shutdown")
		}
	}()
	g.Go(func() error {
		log.Printf("Server Launched %s ...\n", *addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
}

func collect(wg *sync.WaitGroup, ctx context.Context, client *http.Client, kmurl string) {
	defer wg.Done()

	// Kafka Manager API spec https://github.com/yahoo/kafka-manager/blob/master/conf/routes
	apiClusterURL := fmt.Sprintf("%s/api/status/clusters", kmurl)
	var clusters kafka_manager_exporter.APIClusters
	call(ctx, client, apiClusterURL, &clusters)

	var clusterWG sync.WaitGroup
	for _, c := range clusters.Clusters.Active {
		clusterWG.Add(1)
		go func(clusterName string) {
			brokersHTMLURL := fmt.Sprintf("%s/clusters/%s/brokers", kmurl, clusterName)
			callHTML(ctx, client, brokersHTMLURL, "table", func(selection *goquery.Selection) {
				selection.First().Find("tr").EachWithBreak(func(i int, selection *goquery.Selection) bool {
					c := selection.Children()
					if i == 0 {
						if c.Eq(1).Text() != "Host" {
							return false
						}
						return true
					}

					host := strings.TrimSpace(c.Eq(1).Text())
					brokerBytesIn.WithLabelValues(clusterName, host).Set(parseRateFormat(strings.TrimSpace(c.Eq(4).Text())))
					brokerBytesOut.WithLabelValues(clusterName, host).Set(parseRateFormat(strings.TrimSpace(c.Eq(5).Text())))
					return true
				})
				selection.Last().Find("tr").EachWithBreak(func(i int, selection *goquery.Selection) bool {
					c := selection.Children()
					if i == 0 {
						if c.Eq(0).Text() != "Rate" {
							return false
						}
						return true
					}

					labelOfRow := strings.TrimSpace(c.Eq(0).Text())
					if len(labelOfRow) == 0 {
						// goquery's bug can't handle tr correctly
						return true
					}
					matrics, ok := labelToClusterMetrics[labelOfRow]
					if !ok {
						log.Printf("unknown label '%s'", labelOfRow)
						return true
					}
					matrics.WithLabelValues(clusterName).Set(parseRateFormat(strings.TrimSpace(c.Eq(2).Text())))
					return true
				})
			})

			apiTopicidentiesURL := fmt.Sprintf("%s/api/status/%s/topicIdentities", kmurl, clusterName)
			var topicidenties kafka_manager_exporter.APITopicIdentities
			call(ctx, client, apiTopicidentiesURL, &topicidenties)
			for _, t := range topicidenties.TopicIdentities {
				topicSummedOffset.WithLabelValues(clusterName, t.Topic).Set(float64(t.SummedTopicOffsets))
				rateSum := float64(0)
				for _, tp := range t.PartitionsIdentity {
					topicOffset.WithLabelValues(clusterName, t.Topic, strconv.Itoa(tp.PartNum)).Set(float64(tp.LatestOffset))
					topicOffset.WithLabelValues(clusterName, t.Topic, strconv.Itoa(tp.PartNum)).Set(float64(tp.LatestOffset))
					topicPerPartitionChangeRate.WithLabelValues(clusterName, t.Topic, strconv.Itoa(tp.PartNum)).Set(tp.RateOfChange)
					rateSum += tp.RateOfChange
				}
				topicChangeRate.WithLabelValues(clusterName, t.Topic).Set(rateSum)


				_ = `
				topicHTMLURL := fmt.Sprintf("%s/clusters/%s/topics/%s", kmurl, clusterName, t.Topic)
				callHTML(ctx, client, topicHTMLURL, "table:nth-child(2)", func(selection *goquery.Selection) {
					selection.Last().Find("tr").EachWithBreak(func(i int, selection *goquery.Selection) bool {
						c := selection.Children()
						if i == 0 {
							if c.Eq(0).Text() != "Rate" {
								return false
							}
							return true
						}

						labelOfRow := strings.TrimSpace(c.Eq(0).Text())
						if len(labelOfRow) == 0 {
							// goquery's bug can't handle tr correctly
							return true
						}
						matrics, ok := labelToTopicMetrics[labelOfRow]
						if !ok {
							log.Printf("unknown label '%s'", labelOfRow)
							return true
						}
						matrics.WithLabelValues(clusterName).Set(parseRateFormat(strings.TrimSpace(c.Eq(2).Text())))
						return true
					})
				})
				`
			}

			apiConsumerSummaryURL := fmt.Sprintf("%s/api/status/%s/consumersSummary", kmurl, clusterName)
			var consumers kafka_manager_exporter.APIConsumers
			call(ctx, client, apiConsumerSummaryURL, &consumers)

			var consumerWG sync.WaitGroup
			for _, consumer := range consumers.Consumers {
				for _, consumerTopic := range consumer.Topics {
					consumerWG.Add(1)
					go func(clusterName, cName, cTopic, cType string) {
						apiConsumerTopicSummaryURL := fmt.Sprintf("%s/api/status/%s/%s/%s/%s/topicSummary", kmurl, clusterName, cName, cTopic, cType)
						var consumerTopicSummary kafka_manager_exporter.APITopicSummary
						call(ctx, client, apiConsumerTopicSummaryURL, &consumerTopicSummary)

						consumerTopicLag.WithLabelValues(clusterName, cName, cTopic).Set(float64(consumerTopicSummary.TotalLag)) consumerTopicPercentageCovered.WithLabelValues(clusterName, cName, cTopic).Set(float64(consumerTopicSummary.PercentageCovered))
						for n, partition := range consumerTopicSummary.PartitionOffsets {
							consumerTopicPartitionOffset.WithLabelValues(clusterName, cName, cTopic, strconv.Itoa(n)).Set(float64(partition))
						}
						for n, partition := range consumerTopicSummary.PartitionLatestOffsets {
							consumerTopicPartitionLatestOffset.WithLabelValues(clusterName, cName, cTopic, strconv.Itoa(n)).Set(float64(partition))
						}
						consumerWG.Done()
					}(clusterName, consumer.Name, consumerTopic, consumer.Type)
				}
			}
			consumerWG.Wait()
			clusterWG.Done()
		}(c.Name)
	}
	clusterWG.Wait()
}

func call(ctx context.Context, client *http.Client, url string, val interface{}) {
	v := time.Now()
	defer rpcDurations.WithLabelValues().Observe(time.Now().Sub(v).Seconds())
	req, _ := http.NewRequest("GET", url, nil)
	req = req.WithContext(ctx)
	if resp, err := client.Do(req); err != nil && strings.Contains(err.Error(), context.Canceled.Error()) {
		log.Printf("fetch %s failed: %s", url, err)
		return
	} else {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(val); err != nil {
			log.Printf("decoding cluster status failed: %s", err)
			return
		}
	}
}

func callHTML(ctx context.Context, client *http.Client, pageUrl, selector string, acquireFunc func(selection *goquery.Selection)) {
	v := time.Now()
	defer rpcDurations.WithLabelValues().Observe(time.Now().Sub(v).Seconds())

	req, _ := http.NewRequest("GET", pageUrl, nil)
	req = req.WithContext(ctx)
	if resp, err := client.Do(req); err != nil && strings.Contains(err.Error(), context.Canceled.Error()) {
		log.Printf("fetch %s failed: %s", pageUrl, err)
		return
	} else {
		defer resp.Body.Close()
		if doc, err := goquery.NewDocumentFromReader(resp.Body); err != nil {
			log.Printf("parsing broker page failed: %s", err)
		} else {
			sel := doc.Find(selector)
			acquireFunc(sel)
		}
	}

}

func parseRateFormat(rate string) float64 {
	// https://github.com/yahoo/kafka-manager/blob/master/app/kafka/manager/jmx/KafkaJMX.scala#L381
	lrate := len(rate)
	if lrate == 0 {
		return 0
	}
	for n, i := range []byte("kmbt") {
		if rate[lrate-1] == i {
			rateNum, err := strconv.ParseFloat(rate[:lrate-1], 64)
			if err != nil {
				return 0
			}
			return rateNum * math.Pow(1000, float64(n+1))
		}
	}
	return 0
}
