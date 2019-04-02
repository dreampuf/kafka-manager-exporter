package kafka_manager_exporter

type APIClusters struct {
	Clusters struct {
		Active []struct {
			Name          string `json:"name"`
			CuratorConfig struct {
				ZkConnect       string `json:"zkConnect"`
				ZkMaxRetry      int    `json:"zkMaxRetry"`
				BaseSleepTimeMs int    `json:"baseSleepTimeMs"`
				MaxSleepTimeMs  int    `json:"maxSleepTimeMs"`
			} `json:"curatorConfig"`
			Enabled bool `json:"enabled"`
			Version struct {
			} `json:"version"`
			JmxEnabled               bool `json:"jmxEnabled"`
			JmxSsl                   bool `json:"jmxSsl"`
			PollConsumers            bool `json:"pollConsumers"`
			FilterConsumers          bool `json:"filterConsumers"`
			LogkafkaEnabled          bool `json:"logkafkaEnabled"`
			ActiveOffsetCacheEnabled bool `json:"activeOffsetCacheEnabled"`
			DisplaySizeEnabled       bool `json:"displaySizeEnabled"`
			Tuning                   struct {
				BrokerViewUpdatePeriodSeconds       int `json:"brokerViewUpdatePeriodSeconds"`
				ClusterManagerThreadPoolSize        int `json:"clusterManagerThreadPoolSize"`
				ClusterManagerThreadPoolQueueSize   int `json:"clusterManagerThreadPoolQueueSize"`
				KafkaCommandThreadPoolSize          int `json:"kafkaCommandThreadPoolSize"`
				KafkaCommandThreadPoolQueueSize     int `json:"kafkaCommandThreadPoolQueueSize"`
				LogkafkaCommandThreadPoolSize       int `json:"logkafkaCommandThreadPoolSize"`
				LogkafkaCommandThreadPoolQueueSize  int `json:"logkafkaCommandThreadPoolQueueSize"`
				LogkafkaUpdatePeriodSeconds         int `json:"logkafkaUpdatePeriodSeconds"`
				PartitionOffsetCacheTimeoutSecs     int `json:"partitionOffsetCacheTimeoutSecs"`
				BrokerViewThreadPoolSize            int `json:"brokerViewThreadPoolSize"`
				BrokerViewThreadPoolQueueSize       int `json:"brokerViewThreadPoolQueueSize"`
				OffsetCacheThreadPoolSize           int `json:"offsetCacheThreadPoolSize"`
				OffsetCacheThreadPoolQueueSize      int `json:"offsetCacheThreadPoolQueueSize"`
				KafkaAdminClientThreadPoolSize      int `json:"kafkaAdminClientThreadPoolSize"`
				KafkaAdminClientThreadPoolQueueSize int `json:"kafkaAdminClientThreadPoolQueueSize"`
			} `json:"tuning"`
			SecurityProtocol struct {
			} `json:"securityProtocol"`
		} `json:"active"`
		Pending []interface{} `json:"pending"`
	} `json:"clusters"`
}

type APITopicIdentities struct {
	TopicIdentities []struct {
		Topic              string `json:"topic"`
		ReadVersion        int    `json:"readVersion"`
		Partitions         int    `json:"partitions"`
		PartitionsIdentity []struct {
			PartNum           int     `json:"partNum"`
			Leader            int     `json:"leader"`
			LatestOffset      int     `json:"latestOffset"`
			RateOfChange      float64 `json:"rateOfChange"`
			Isr               []int   `json:"isr"`
			Replicas          []int   `json:"replicas"`
			IsPreferredLeader bool    `json:"isPreferredLeader"`
			IsUnderReplicated bool    `json:"isUnderReplicated"`
		} `json:"partitionsIdentity"`
		NumBrokers        int        `json:"numBrokers"`
		ConfigReadVersion int        `json:"configReadVersion"`
		Config            [][]string `json:"config"`
		ClusterContext    struct {
			ClusterFeatures struct {
				Features []struct {
				} `json:"features"`
			} `json:"clusterFeatures"`
			Config struct {
				Name          string `json:"name"`
				CuratorConfig struct {
					ZkConnect       string `json:"zkConnect"`
					ZkMaxRetry      int    `json:"zkMaxRetry"`
					BaseSleepTimeMs int    `json:"baseSleepTimeMs"`
					MaxSleepTimeMs  int    `json:"maxSleepTimeMs"`
				} `json:"curatorConfig"`
				Enabled bool `json:"enabled"`
				Version struct {
				} `json:"version"`
				JmxEnabled               bool `json:"jmxEnabled"`
				JmxSsl                   bool `json:"jmxSsl"`
				PollConsumers            bool `json:"pollConsumers"`
				FilterConsumers          bool `json:"filterConsumers"`
				LogkafkaEnabled          bool `json:"logkafkaEnabled"`
				ActiveOffsetCacheEnabled bool `json:"activeOffsetCacheEnabled"`
				DisplaySizeEnabled       bool `json:"displaySizeEnabled"`
				Tuning                   struct {
					BrokerViewUpdatePeriodSeconds       int `json:"brokerViewUpdatePeriodSeconds"`
					ClusterManagerThreadPoolSize        int `json:"clusterManagerThreadPoolSize"`
					ClusterManagerThreadPoolQueueSize   int `json:"clusterManagerThreadPoolQueueSize"`
					KafkaCommandThreadPoolSize          int `json:"kafkaCommandThreadPoolSize"`
					KafkaCommandThreadPoolQueueSize     int `json:"kafkaCommandThreadPoolQueueSize"`
					LogkafkaCommandThreadPoolSize       int `json:"logkafkaCommandThreadPoolSize"`
					LogkafkaCommandThreadPoolQueueSize  int `json:"logkafkaCommandThreadPoolQueueSize"`
					LogkafkaUpdatePeriodSeconds         int `json:"logkafkaUpdatePeriodSeconds"`
					PartitionOffsetCacheTimeoutSecs     int `json:"partitionOffsetCacheTimeoutSecs"`
					BrokerViewThreadPoolSize            int `json:"brokerViewThreadPoolSize"`
					BrokerViewThreadPoolQueueSize       int `json:"brokerViewThreadPoolQueueSize"`
					OffsetCacheThreadPoolSize           int `json:"offsetCacheThreadPoolSize"`
					OffsetCacheThreadPoolQueueSize      int `json:"offsetCacheThreadPoolQueueSize"`
					KafkaAdminClientThreadPoolSize      int `json:"kafkaAdminClientThreadPoolSize"`
					KafkaAdminClientThreadPoolQueueSize int `json:"kafkaAdminClientThreadPoolQueueSize"`
				} `json:"tuning"`
				SecurityProtocol struct {
				} `json:"securityProtocol"`
			} `json:"config"`
		} `json:"clusterContext"`
		Size               interface{} `json:"size"`
		ReplicationFactor  int         `json:"replicationFactor"`
		PartitionsByBroker []struct {
			ID             int   `json:"id"`
			Partitions     []int `json:"partitions"`
			IsSkewed       bool  `json:"isSkewed"`
			Leaders        []int `json:"leaders"`
			IsLeaderSkewed bool  `json:"isLeaderSkewed"`
		} `json:"partitionsByBroker"`
		SummedTopicOffsets          int64  `json:"summedTopicOffsets"`
		PreferredReplicasPercentage int    `json:"preferredReplicasPercentage"`
		UnderReplicatedPercentage   int    `json:"underReplicatedPercentage"`
		TopicBrokers                int    `json:"topicBrokers"`
		BrokersSkewPercentage       int    `json:"brokersSkewPercentage"`
		BrokersSpreadPercentage     int    `json:"brokersSpreadPercentage"`
		ProducerRate                string `json:"producerRate"`
	} `json:"topicIdentities"`
}

type APITopicSummary struct {
	TotalLag               int64    `json:"totalLag"`
	PercentageCovered      int      `json:"percentageCovered"`
	PartitionOffsets       []int64  `json:"partitionOffsets"`
	PartitionLatestOffsets []int64  `json:"partitionLatestOffsets"`
	Owners                 []string `json:"owners"`
}

type APIConsumers struct {
	Consumers []struct {
		Name   string   `json:"name"`
		Type   string   `json:"type"`
		Topics []string `json:"topics"`
	} `json:"consumers"`
}
