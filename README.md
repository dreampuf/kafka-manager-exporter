# kafka-manager-exporter
A Prometheus Exporter For Kafka Manager

# Usage 

```bash
$ kafka-manager-exporter --help
  -addr string
    	http server listening address (default "127.0.0.1:9001")
  -interval duration
    	interval of collecting (default 15s)
  -urls string
    	Kafka Manager URLs. Separate with comma (default "http://127.0.0.1:9000")

$ kafka-manager-exporter -urls http://KAFKA_MANAGER_URL -interval 15s -addr 0.0.0.1:9001

2019/04/02 17:56:00 Server Launched 127.0.0.1:9001 ...
^C2019/04/02 17:56:18 Server Shutdown

# If you prefer docker

$ docker run -it --rm -p 9001:9001 dreampuf/kafka-manager-exporter -urls http://KAFKA_MANAGER_URL -interval 15s -addr 0.0.0.0:9001
2019/04/02 22:28:03 Server Launched 0.0.0.0:9001 ...
```

```bash
$ curl localhost:9001/metrics
topic_summed_offset{topic="pick_ack_pair"} 3.0283061e+07
topic_summed_offset{topic="test"} 8.080727e+06
topic_summed_offset{topic="test1"} 1.087054e+06
topic_summed_offset{topic="test2"} 45
topic_summed_offset{topic="test_broker"} 108723
topic_summed_offset{topic="test_new_consumer"} 1.0209415389e+10
topic_summed_offset{topic="test_performance"} 79514
topic_summed_offset{topic="test_single_partition"} 4.725704e+06
topic_summed_offset{topic="test_topic"} 41021
```

# Why

We already have a [kafka-exporter](https://github.com/danielqsj/kafka_exporter). But we always have multiple kafka cluster to manage. kafka manager becomes a kind of must have component. Rather than setup a batch of indivitually kafka-exporter instances, I can gather these data in the center place (Yes, it's break your best practice. Brain).

PLUS, if you also want some bandwidth metrics, you may need to set up [jmx-exporter](https://github.com/prometheus/jmx_exporter) as well. That's the efforts.

All these ideas come to this project. PR are welcome.

Author: Dreampuf
