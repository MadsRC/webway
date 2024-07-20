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

	if opts.Datastore == nil {
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
	Datastore           metadatastore.Datastore
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

func WithMetadataStoreServerDatastore(datastore metadatastore.Datastore) MetadataStoreServerOption {
	return newFuncMetadataStoreServerOption(func(o *metadataStoreServerOptions) {
		o.Datastore = datastore
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
	err := m.opts.Datastore.CreateAgent(ctx, &metadatastore.Agent{
		ID:               request.AgentId,
		AvailabilityZone: request.AvailabilityZone,
	})
	if err != nil && !errors.Is(err, pgx.ErrAlreadyExists) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrAlreadyExists) {
		err = m.opts.Datastore.UpdateAgent(ctx, &metadatastore.Agent{
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
	err := m.opts.Datastore.DeleteAgent(ctx, request.AgentId)
	if err != nil {
		return nil, err
	}
	return &pb.DeregisterAgentResponse{}, nil
}

func (m *MetadataStoreServer) GetMetadata(ctx context.Context, request *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	var response pb.GetMetadataResponse

	az := metadatastore.AvailabilityZone(request.AvailabilityZone)

	_, agents, leaderIndex, err := az.Healthy(ctx, 1, m.opts.AgentAliveThreshold, m.opts.Datastore)
	if err != nil {
		return nil, err
	}

	response.ClusterId = m.opts.ClusterID
	if agents != nil {
		response.Agents = make([]*pb.Agent, len(agents))
		for i, agent := range agents {
			response.Agents[i] = &pb.Agent{
				Id:               agent.ID,
				AvailabilityZone: agent.AvailabilityZone,
			}
		}
	}

	topicMap := make(map[string]*pb.Topic)

	topics, err := m.opts.Datastore.ReadAllTopics(ctx)
	if err != nil {
		return nil, err
	}
	partitions, err := m.opts.Datastore.ReadAllPartitions(ctx)
	if err != nil {
		return nil, err
	}

	for _, partition := range partitions {
		if topicMap[partition.TopicID] == nil {
			topicMap[partition.TopicID] = &pb.Topic{
				TopicId:    partition.TopicID,
				Partitions: make([]*pb.Partition, 0),
			}
		}
		topicMap[partition.TopicID].Partitions = append(topicMap[partition.TopicID].Partitions, &pb.Partition{
			LeaderId: agents[leaderIndex].ID,
			Id:       int32(partition.ID),
		})
	}

	for _, topic := range topics {
		if topicMap[topic.ID] != nil {
			topicMap[topic.ID].Name = topic.Name
		}
	}

	for _, topic := range topicMap {
		response.Topics = append(response.Topics, topic)
	}

	return &response, nil
}
