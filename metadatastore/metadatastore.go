package metadatastore

import (
	"context"
	"time"
)

type Agent struct {
	ID               string    `db:"agent_id"`
	AvailabilityZone string    `db:"availability_zone"`
	LastSeen         time.Time `db:"last_seen"`
}

func (a *Agent) IsLive(threshold time.Duration) bool {
	return time.Since(a.LastSeen) < threshold
}

const (
	NoAz = AvailabilityZone("")
)

type AvailabilityZone string

func (az AvailabilityZone) String() string {
	return string(az)
}

// Healthy returns true if the availability zone has a quorum of live agents.
//
// The agents returned by this function are considered live within the availability zone, as they will have sent a
// heartbeat within the defined livenessThreshold.
//
// If the availability zone is empty (i.e. NoAz), then all agents that have sent a heartbeat within the defined
// livenessThreshold are considered live, regardless of their availability zone.
func (az AvailabilityZone) Healthy(ctx context.Context, quorum int, livenessThreshold time.Duration, ds Datastore) (bool, []*Agent, error) {
	allAgents, err := ds.ReadAllAgents(ctx)
	if err != nil {
		return false, nil, err
	}

	liveAgents := make([]*Agent, 0)
	for _, agent := range allAgents {
		if agent.IsLive(livenessThreshold) {
			if agent.AvailabilityZone == string(az) || az == NoAz {
				liveAgents = append(liveAgents, agent)
			}
		}
	}

	if len(liveAgents) >= quorum {
		return true, liveAgents, nil
	}
	return false, nil, nil
}

type Datastore interface {
	CreateAgent(ctx context.Context, agent *Agent) error
	ReadAgent(ctx context.Context, agentID string) (*Agent, error)
	ReadAllAgents(ctx context.Context) ([]*Agent, error)
	UpdateAgent(ctx context.Context, agent *Agent) error
	DeleteAgent(ctx context.Context, agentID string) error

	CreateTopic(ctx context.Context, topic *Topic) error
	ReadTopic(ctx context.Context, topicID string) (*Topic, error)
	ReadAllTopics(ctx context.Context) ([]*Topic, error)
	UpdateTopic(ctx context.Context, topic *Topic) error
	DeleteTopic(ctx context.Context, topicID string) error

	CreatePartition(ctx context.Context, partition *Partition) error
	ReadPartition(ctx context.Context, partitionID string) (*Partition, error)
	ReadAllPartitions(ctx context.Context) ([]*Partition, error)
	UpdatePartition(ctx context.Context, partition *Partition) error
	DeletePartition(ctx context.Context, partitionID string) error
}

type Topic struct {
	// The ID is a UUIDv4 to keep inline with KIP-516
	ID       string `db:"topic_id"`
	Name     string `db:"name"`
	Internal bool   `db:"internal"`
}

type Partition struct {
	TopicID string `db:"topic_id"`
	ID      int    `db:"partition_id"`
}

type UUIDService interface {
	RandomUUID() (string, error)
	ParseUUID(uuid string) ([]byte, error)
	FormatUUID(uuid []byte) (string, error)
}
