package main

import (
	"context"
	"github.com/madsrc/webway"
	"github.com/madsrc/webway/agent/internal/grpc"
	"github.com/madsrc/webway/agent/internal/sarama"
	pb "github.com/madsrc/webway/gen/go/webway/v1"
	"github.com/madsrc/webway/koanf"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type services struct {
	metadataStoreGrpcConn   *grpc.ClientConn
	metadataStoreGrpcClient pb.MetadataStoreClient
	kafkaServer             *sarama.KafkaServer
}

func (s *services) GracefulStop() {
	s.kafkaServer.GracefulStop()
	s.metadataStoreGrpcConn.GracefulStop()
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := koanf.NewConfig(
		koanf.WithConfigNestedMap(DefaultConfig),
	)
	if err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	svcs, err := setupServices(cfg)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	advertisedHostname := strings.Split(cfg.String("agent.advertised_address"), ":")[0]
	advertisedPort := 9092
	if len(strings.Split(cfg.String("agent.advertised_address"), ":")) > 1 {
		advertisedPort, err = strconv.Atoi(strings.Split(cfg.String("advertised_address"), ":")[1])
		if err != nil {
			log.Fatalf("failed to parse port from agent.advertised_address: %v", err)
		}
	}

	// Registering the agent
	raresp, err := svcs.metadataStoreGrpcClient.RegisterAgent(ctx, &pb.RegisterAgentRequest{
		Hostname:         advertisedHostname,
		Port:             int32(advertisedPort),
		AvailabilityZone: cfg.String("agent.availability_zone"),
	})
	if err != nil {
		log.Fatalf("failed to register agent: %v", err)
	}

	hbInterval, err := time.ParseDuration(cfg.String("metadatastore.heartbeat_interval"))
	if err != nil {
		log.Fatalf("failed to parse heartbeat interval: %v", err)
	}
	hbTicker := time.NewTicker(hbInterval)
	hbQuit := make(chan struct{})

	lis, err := net.Listen("tcp", cfg.String("agent.listen_address"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		for {
			select {
			case <-hbTicker.C:
				heartbeat(ctx, raresp.Id, svcs)
			case <-hbQuit:
				return
			}
		}
	}()

	go func() {
		_ = svcs.kafkaServer.Serve(lis)
	}()

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Print("received shutdown signal")

	// Deregistering the agent
	_, err = svcs.metadataStoreGrpcClient.DeregisterAgent(ctx, &pb.DeregisterAgentRequest{
		Id: raresp.Id,
	})
	if err != nil {
		log.Fatalf("failed to deregister agent: %v", err)
	}

	cancel()
	svcs.GracefulStop()
}

func setupServices(cfg webway.Config) (*services, error) {
	svcs := &services{}
	var err error

	svcs.metadataStoreGrpcConn, err = grpc.NewClientConn(cfg.String("metadatastore.grpc.address"), grpc.WithClientConnInsecure())
	if err != nil {
		return nil, err
	}

	svcs.metadataStoreGrpcClient = pb.NewMetadataStoreClient(svcs.metadataStoreGrpcConn)

	svcs.kafkaServer, err = sarama.NewKafkaServer(
		sarama.WithMetadataStoreGrpcClient(svcs.metadataStoreGrpcClient),
	)

	return svcs, nil
}

var DefaultConfig = map[string]interface{}{
	"metadatastore.grpc.address":       "127.0.0.1:9068",
	"metadatastore.heartbeat_interval": "3s",
	"agent.listen_address":             "0.0.0.0:9092",
	"agent.advertised_address":         "127.0.0.1:9092",
	"agent.availability_zone":          "us-east-1",
}

func heartbeat(ctx context.Context, id int32, svcs *services) {
	_, err := svcs.metadataStoreGrpcClient.AgentHeartbeat(ctx, &pb.Heartbeat{
		Id: id,
	})
	if err != nil {
		log.Printf("failed to send heartbeat: %v", err)
	}
}
