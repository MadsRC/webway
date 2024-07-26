package main

import (
	"flag"
	"fmt"
	"github.com/MadsRC/sarama"
	"log"
	"strings"
)

func main() {
	var brokerAddr string
	var clientID string
	var version int
	var topicsStr string
	var allowAutoTopicCreation bool
	var includeClusterAuthorizedOperations bool
	var includeTopicAuthorizedOperations bool

	flag.StringVar(&brokerAddr, "brokerAddr", "localhost:9092", "brokerAddr to connect to")
	flag.StringVar(&clientID, "clientID", "kafka-get-metadata", "clientID to use")
	flag.IntVar(&version, "version", 1, "version of the metadata request")
	flag.StringVar(&topicsStr, "topics", "", "comma separated list of topics to request")
	flag.BoolVar(&allowAutoTopicCreation, "allowAutoTopicCreation", false, "allowAutoTopicCreation")
	flag.BoolVar(&includeClusterAuthorizedOperations, "includeClusterAuthorizedOperations", false, "includeClusterAuthorizedOperations")
	flag.BoolVar(&includeTopicAuthorizedOperations, "includeTopicAuthorizedOperations", false, "includeTopicAuthorizedOperations")

	flag.Parse()

	broker := sarama.NewBroker(brokerAddr)
	defer broker.Close()

	cfg := sarama.NewConfig()
	cfg.ClientID = clientID
	cfg.Net.MaxOpenRequests = 1

	err := broker.Open(cfg)
	if err != nil {
		log.Fatalf("error opening connection to broker '%s': %s", brokerAddr, err)
	}

	connected, err := broker.Connected()
	if err != nil {
		log.Fatalf("error getting connection status from broker '%s': %s", brokerAddr, err)
	}

	if !connected {
		log.Fatalf("broker '%s' is not connected", brokerAddr)
	}

	mdReq := sarama.MetadataRequest{}
	mdReq.AllowAutoTopicCreation = allowAutoTopicCreation
	mdReq.IncludeClusterAuthorizedOperations = includeClusterAuthorizedOperations
	mdReq.IncludeTopicAuthorizedOperations = includeTopicAuthorizedOperations
	if topicsStr != "" {
		mdReq.Topics = strings.Split(topicsStr, ",")
	}
	mdReq.Version = int16(version)

	fmt.Printf("Request:\n%+v\n", mdReq)

	metadata, err := broker.GetMetadata(&mdReq)
	if err != nil {
		log.Fatalf("error getting metadata from broker '%s': %s", brokerAddr, err)
	}

	fmt.Printf("Response:\n%+v\n", metadata)
	for _, topic := range metadata.Topics {
		fmt.Printf("Topic: %+v\n", topic)
		for _, partition := range topic.Partitions {
			fmt.Printf("Partition: %+v\n", partition)
		}
	}
	for _, broker := range metadata.Brokers {
		fmt.Printf("Broker: %+v\n", broker)
	}
}
