package main

import (
	"context"
	"github.com/madsrc/webway"
	"github.com/madsrc/webway/agent/internal/grpc"
	pb "github.com/madsrc/webway/gen/go/webway/v1"
	"github.com/madsrc/webway/koanf"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type services struct {
	metadataStoreGrpcConn   *grpc.ClientConn
	metadataStoreGrpcClient pb.MetadataStoreClient
}

func (s *services) GracefulStop() {
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

	_, err = svcs.metadataStoreGrpcClient.RegisterAgent(ctx, &pb.RegisterAgentRequest{
		AgentId:          "agent-1",
		AvailabilityZone: "us-east-1",
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

	go func() {
		for {
			select {
			case <-hbTicker.C:
				heartbeat(ctx, svcs)
			case <-hbQuit:
				return
			}
		}
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
		AgentId: "agent-1",
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

	return svcs, nil
}

var DefaultConfig = map[string]interface{}{
	"metadatastore.grpc.address":       "127.0.0.1:9068",
	"metadatastore.heartbeat_interval": "3s",
}

func heartbeat(ctx context.Context, svcs *services) {
	_, err := svcs.metadataStoreGrpcClient.RegisterAgent(ctx, &pb.RegisterAgentRequest{
		AgentId:          "agent-1",
		AvailabilityZone: "us-east-1",
	})
	if err != nil {
		log.Printf("failed to send heartbeat: %v", err)
	}
}
