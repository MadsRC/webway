package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/madsrc/webway"
	pb "github.com/madsrc/webway/gen/go/webway/v1"
	"github.com/madsrc/webway/metadatastore"
	"github.com/madsrc/webway/metadatastore/internal/pgx"
	"time"
)

type MetadataStoreServer struct {
	opts *metadataStoreServerOptions
	pb.UnimplementedMetadataStoreServer
}

func NewMetadataStoreServer(opt ...MetadataStoreServerOption) (*MetadataStoreServer, error) {
	opts := defaultMetadataStoreServerOptions
	for _, o := range globalMetadataStoreServerOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}

	if opts.AgentDatastore == nil {
		return nil, &webway.MissingOptionError{Option: "agentdatastore"}
	}
	if opts.UUIDService == nil {
		return nil, &webway.MissingOptionError{Option: "uuidservice"}
	}

	mds := &MetadataStoreServer{
		opts: &opts,
	}

	_, err := mds.generateClusterID()
	if err != nil {
		return nil, err
	}

	return mds, nil
}

type metadataStoreServerOptions struct {
	AgentDatastore      metadatastore.Datastore
	AgentAliveThreshold time.Duration
	UUIDService         metadatastore.UUIDService
	ClusterID           string
}

var defaultMetadataStoreServerOptions = metadataStoreServerOptions{
	AgentAliveThreshold: 10 * time.Second,
}
var globalMetadataStoreServerOptions []MetadataStoreServerOption

type MetadataStoreServerOption interface {
	apply(*metadataStoreServerOptions)
}

// funcMetadataStoreServerOption wraps a function that modifies metadataStoreServerOption into an
// implementation of the MetadataStoreServerOption interface.
type funcMetadataStoreServerOption struct {
	f func(*metadataStoreServerOptions)
}

func (fdo *funcMetadataStoreServerOption) apply(do *metadataStoreServerOptions) {
	fdo.f(do)
}

func newFuncMetadataStoreServerOption(f func(*metadataStoreServerOptions)) *funcMetadataStoreServerOption {
	return &funcMetadataStoreServerOption{
		f: f,
	}
}

func WithMetadataStoreServerAgentStore(agentStore metadatastore.Datastore) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.AgentDatastore = agentStore
	})
}

func WithMetadataStoreServerAgentAliveThreshold(agentAliveThreshold time.Duration) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.AgentAliveThreshold = agentAliveThreshold
	})
}

func WithMetadataStoreServerUUIDService(uuidService metadatastore.UUIDService) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.UUIDService = uuidService
	})
}

func WithMetadataStoreServerClusterID(clusterID string) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.ClusterID = clusterID
	})
}

func (m *MetadataStoreServer) generateClusterID() (string, error) {
	if m.opts.ClusterID != "" {
		return m.opts.ClusterID, nil
	}

	str, err := m.opts.UUIDService.RandomUUID()
	if err != nil {
		return "", err
	}
	b, err := m.opts.UUIDService.ParseUUID(str)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (m *MetadataStoreServer) RegisterAgent(ctx context.Context, request *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	err := m.opts.AgentDatastore.CreateAgent(ctx, &metadatastore.Agent{
		ID:               request.AgentId,
		AvailabilityZone: request.AvailabilityZone,
	})
	if err != nil && !errors.Is(err, pgx.ErrAlreadyExists) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrAlreadyExists) {
		err = m.opts.AgentDatastore.UpdateAgent(ctx, &metadatastore.Agent{
			ID:               request.AgentId,
			AvailabilityZone: request.AvailabilityZone,
		})
		if err != nil {
			return nil, err
		}
	}
	return &pb.RegisterAgentResponse{}, nil
}

func (m *MetadataStoreServer) DeregisterAgent(ctx context.Context, request *pb.DeregisterAgentRequest) (*pb.DeregisterAgentResponse, error) {
	err := m.opts.AgentDatastore.DeleteAgent(ctx, request.AgentId)
	if err != nil {
		return nil, err
	}
	return &pb.DeregisterAgentResponse{}, nil
}

func (m *MetadataStoreServer) GetMetadata(ctx context.Context, request *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	var response pb.GetMetadataResponse

	az := metadatastore.AvailabilityZone(request.AvailabilityZone)

	_, agents, err := az.Healthy(ctx, 1, m.opts.AgentAliveThreshold, m.opts.AgentDatastore)
	if err != nil {
		return nil, err
	}

	response.ClusterId = m.opts.ClusterID
	response.Agents = make([]*pb.Agent, 0)

	for _, agent := range agents {
		response.Agents = append(response.Agents, &pb.Agent{
			Id:               agent.ID,
			AvailabilityZone: agent.AvailabilityZone,
		})
	}

	// Next step is to return the list of topics, and select an agent/broker to be the leader for all topic partitions
	// as per https://www.warpstream.com/blog/hacking-the-kafka-protocol#load-balancing
	// This needs to be done in a round-robin fashion... Which necessitates keeping some state.

	return &response, nil
}
