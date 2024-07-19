package main

import (
	"context"
	"errors"
	"github.com/madsrc/webway"
	"github.com/madsrc/webway/koanf"
	"github.com/madsrc/webway/metadatastore"
	"github.com/madsrc/webway/metadatastore/internal/grpc"
	"github.com/madsrc/webway/metadatastore/internal/migrate"
	"github.com/madsrc/webway/metadatastore/internal/pgx"
	"github.com/madsrc/webway/metadatastore/internal/uuid"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type services struct {
	migrationService        *migrate.MigrationService
	agentDatastore          *pgx.Datastore
	grpcMetadataStoreServer *grpc.MetadataStoreServer
	grpcServer              *grpc.Server
	uuidService             metadatastore.UUIDService
}

func (s *services) GracefulStop() {
	s.grpcServer.GracefulStop()
	s.agentDatastore.GracefulStop()
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

	// run database migrations
	err = svcs.migrationService.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to run migrations: %v", err)
	}

	lis, err := net.Listen("tcp", cfg.String("grpc.listen_address"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	go func() {
		svcs.grpcServer.Serve(lis)
	}()

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Print("received shutdown signal")

	cancel()
	svcs.GracefulStop()
}

func setupServices(cfg webway.Config) (*services, error) {
	svcs := &services{}
	var err error

	svcs.migrationService, err = migrate.NewMigrationService(
		migrate.WithMigrationServiceDatabaseConnectionString(cfg.String("datastore.connection_string")),
	)
	if err != nil {
		return nil, err
	}

	svcs.agentDatastore, err = pgx.NewDatastore(
		pgx.WithDatastoreConnectionString(cfg.String("datastore.connection_string")),
	)
	if err != nil {
		return nil, err
	}

	svcs.uuidService, err = uuid.NewUUIDService()
	if err != nil {
		return nil, err
	}

	grpcMdsOptions := []grpc.MetadataStoreServerOption{
		grpc.WithMetadataStoreServerAgentStore(svcs.agentDatastore),
		grpc.WithMetadataStoreServerUUIDService(svcs.uuidService),
	}
	if cfg.String("kafka.cluster_id") != "" {
		grpcMdsOptions = append(grpcMdsOptions, grpc.WithMetadataStoreServerClusterID(cfg.String("kafka.cluster_id")))
	}

	svcs.grpcMetadataStoreServer, err = grpc.NewMetadataStoreServer(
		grpcMdsOptions...,
	)
	if err != nil {
		return nil, err
	}

	svcs.grpcServer, err = setupGrpcServer(cfg, svcs)
	if err != nil {
		return nil, err
	}

	return svcs, nil
}

func setupGrpcServer(cfg webway.Config, svc *services) (*grpc.Server, error) {
	opts := []grpc.ServerOption{
		grpc.WithServerMetadataStoreServer(svc.grpcMetadataStoreServer),
	}
	if cfg.Bool("grpc.reflection.enabled") {
		opts = append(opts, grpc.WithServerReflection())
	} else {
		opts = append(opts, grpc.WithoutServerReflection())
	}

	grpcServer, err := grpc.NewServer(opts...)
	if err != nil {
		return nil, err
	}

	return grpcServer, nil
}

var DefaultConfig = map[string]interface{}{
	"grpc.listen_address":         "127.0.0.1:9068",
	"grpc.reflection.enabled":     true,
	"datastore.connection_string": "postgres://postgres:postgres@localhost:5432/postgres",
}
