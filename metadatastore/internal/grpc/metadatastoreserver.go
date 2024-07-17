package grpc

import (
	"context"
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

	return &MetadataStoreServer{
		opts: &opts,
	}, nil
}

type metadataStoreServerOptions struct {
	AgentDatastore      metadatastore.AgentDatastore
	AgentAliveThreshold time.Duration
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

func WithMetadataStoreServerAgentStore(agentStore metadatastore.AgentDatastore) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.AgentDatastore = agentStore
	})
}

func WithMetadataStoreServerAgentAliveThreshold(agentAliveThreshold time.Duration) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.AgentAliveThreshold = agentAliveThreshold
	})
}

func (m *MetadataStoreServer) RegisterAgent(ctx context.Context, request *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	err := m.opts.AgentDatastore.Create(ctx, &metadatastore.Agent{
		ID:               request.AgentId,
		AvailabilityZone: request.AvailabilityZone,
	})
	if err != nil && !errors.Is(err, pgx.ErrAlreadyExists) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrAlreadyExists) {
		err = m.opts.AgentDatastore.Update(ctx, &metadatastore.Agent{
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
	err := m.opts.AgentDatastore.Delete(ctx, request.AgentId)
	if err != nil {
		return nil, err
	}
	return &pb.DeregisterAgentResponse{}, nil
}

func (m *MetadataStoreServer) GetMetadata(ctx context.Context, request *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	allAgents, err := m.opts.AgentDatastore.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	var response pb.GetMetadataResponse
	response.Agents = make([]*pb.Agent, 0)

	for _, agent := range allAgents {
		if agent.AvailabilityZone == request.AvailabilityZone {
			if time.Since(agent.LastSeen) > m.opts.AgentAliveThreshold {
				continue
			}
			response.Agents = append(response.Agents, &pb.Agent{
				Id:               agent.ID,
				AvailabilityZone: agent.AvailabilityZone,
			})
		}
	}

	return &response, nil
}
