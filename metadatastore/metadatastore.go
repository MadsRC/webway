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

type AgentDatastore interface {
	Create(ctx context.Context, agent *Agent) error
	Read(ctx context.Context, agentID string) (*Agent, error)
	ReadAll(ctx context.Context) ([]*Agent, error)
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, agentID string) error
}
