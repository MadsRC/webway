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
func (az AvailabilityZone) Healthy(ctx context.Context, quorum int, livenessThreshold time.Duration, ds AgentDatastore) (bool, []*Agent, error) {
	allAgents, err := ds.ReadAll(ctx)
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

type AgentDatastore interface {
	Create(ctx context.Context, agent *Agent) error
	Read(ctx context.Context, agentID string) (*Agent, error)
	ReadAll(ctx context.Context) ([]*Agent, error)
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, agentID string) error
}

type UUIDService interface {
	RandomUUID() (string, error)
	ParseUUID(uuid string) ([]byte, error)
	FormatUUID(uuid []byte) (string, error)
}
