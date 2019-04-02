package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dreampuf/kafka-manager-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"log"
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
		[]string{"topic", "partition"},
	)
	topicSummedOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_summed_offset",
			Help: "Topic Summed Offset",
		},
		[]string{"topic"},
	)
	topicChangeRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_change_rate",
			Help: "Topic Change Rate",
		},
		[]string{"topic"},
	)
	topicPerPartitionChangeRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "topic_partition_change_rate",
			Help: "Topic Per Partition Change Rate",
		},
		[]string{"topic", "partition"},
	)
	consumerTopicLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_lag",
			Help: "Consumer Lag Per Topic",
		},
		[]string{"consumer", "topic"},
	)
	consumerTopicPercentageCovered = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_percentage_covered",
			Help: "Consumer Percentage Covered Per Topic",
		},
		[]string{"consumer", "topic"},
	)
	consumerTopicPartitionOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_partition_offset",
			Help: "Consumer Topic Offset Per Partition",
		},
		[]string{"consumer", "topic", "partition"},
	)
	consumerTopicPartitionLatestOffset = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consumer_topic_partition_latest_offset",
			Help: "Consumer Topic Latest Offset Per Partition",
		},
		[]string{"consumer", "topic", "partition"},
	)
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

	apiClusterURL := fmt.Sprintf("%s/api/status/clusters", kmurl)
	var clusters kafka_manager_exporter.APIClusters
	call(ctx, client, apiClusterURL, &clusters)

	var clusterWG sync.WaitGroup
	for _, c := range clusters.Clusters.Active {
		clusterWG.Add(1)
		go func(clusterName string) {
			apiTopicidentiesURL := fmt.Sprintf("%s/api/status/%s/topicIdentities", kmurl, clusterName)
			var topicidenties kafka_manager_exporter.APITopicIdentities
			call(ctx, client, apiTopicidentiesURL, &topicidenties)
			for _, t := range topicidenties.TopicIdentities {
				topicSummedOffset.WithLabelValues(t.Topic).Set(float64(t.SummedTopicOffsets))
				rateSum := float64(0)
				for _, tp := range t.PartitionsIdentity {
					topicOffset.WithLabelValues(t.Topic, strconv.Itoa(tp.PartNum)).Set(float64(tp.LatestOffset))
					topicOffset.WithLabelValues(t.Topic, strconv.Itoa(tp.PartNum)).Set(float64(tp.LatestOffset))
					topicPerPartitionChangeRate.WithLabelValues(t.Topic, strconv.Itoa(tp.PartNum)).Set(tp.RateOfChange)
					rateSum += tp.RateOfChange
				}
				topicChangeRate.WithLabelValues(t.Topic).Set(rateSum)
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

						consumerTopicLag.WithLabelValues(cName, cTopic).Set(float64(consumerTopicSummary.TotalLag))
						consumerTopicPercentageCovered.WithLabelValues(cName, cTopic).Set(float64(consumerTopicSummary.PercentageCovered))
						for n, partition := range consumerTopicSummary.PartitionOffsets {
							consumerTopicPartitionOffset.WithLabelValues(cName, cTopic, strconv.Itoa(n)).Set(float64(partition))
						}
						for n, partition := range consumerTopicSummary.PartitionLatestOffsets {
							consumerTopicPartitionLatestOffset.WithLabelValues(cName, cTopic, strconv.Itoa(n)).Set(float64(partition))
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
		if err := json.NewDecoder(resp.Body).Decode(val); err != nil {
			log.Printf("decoding cluster status failed: %s", err)
			return
		}
	}
}
