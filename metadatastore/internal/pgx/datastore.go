package pgx

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/madsrc/webway"
	"github.com/madsrc/webway/metadatastore"
	"time"
)

type Datastore struct {
	opts datastoreOptions
	db   *pgxpool.Pool
}

func NewDatastore(opt ...DatastoreOption) (*Datastore, error) {
	opts := defaultAgentDatastoreOptions
	for _, o := range globalAgentDatastoreOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	a := &Datastore{opts: opts}

	if a.opts.DatabaseConnectionString == "" {
		return nil, &webway.MissingOptionError{Option: "DatabaseConnectionString"}
	}

	dbpool, err := pgxpool.New(context.Background(), a.opts.DatabaseConnectionString)
	if err != nil {
		return nil, err
	}
	a.db = dbpool

	return a, nil
}

func (a *Datastore) GracefulStop() {
	a.db.Close()
}

func (a *Datastore) CreateAgent(ctx context.Context, agent *metadatastore.Agent) error {
	_, err := a.db.Exec(ctx, "INSERT INTO agents (agent_id, availability_zone, last_seen, deleted_at) VALUES ($1, $2, $3, $4)", agent.ID, agent.AvailabilityZone, time.Now(), nil)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (a *Datastore) ReadAgent(ctx context.Context, agentID string) (*metadatastore.Agent, error) {
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE agent_id = $1 AND deleted_at IS NULL", agentID)
	agent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[metadatastore.Agent])
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (a *Datastore) ReadAllAgents(ctx context.Context) ([]*metadatastore.Agent, error) {
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE deleted_at IS NULL")
	agents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[metadatastore.Agent])
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (a *Datastore) UpdateAgent(ctx context.Context, agent *metadatastore.Agent) error {
	_, err := a.db.Exec(ctx, "UPDATE agents SET availability_zone = $1, last_seen = $2, deleted_at = null WHERE agent_id = $3", agent.AvailabilityZone, time.Now(), agent.ID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Datastore) DeleteAgent(ctx context.Context, agentID string) error {
	tag, err := a.db.Exec(ctx, "UPDATE agents SET deleted_at = $1 WHERE agent_id = $2", time.Now(), agentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *Datastore) CreateTopic(ctx context.Context, topic *metadatastore.Topic) error {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) ReadTopic(ctx context.Context, topicID string) (*metadatastore.Topic, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) ReadAllTopics(ctx context.Context) ([]*metadatastore.Topic, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) UpdateTopic(ctx context.Context, topic *metadatastore.Topic) error {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) DeleteTopic(ctx context.Context, topicID string) error {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) CreatePartition(ctx context.Context, partition *metadatastore.Partition) error {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) ReadPartition(ctx context.Context, partitionID string) (*metadatastore.Partition, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) ReadAllPartitions(ctx context.Context) ([]*metadatastore.Partition, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) UpdatePartition(ctx context.Context, partition *metadatastore.Partition) error {
	//TODO implement me
	panic("implement me")
}

func (a *Datastore) DeletePartition(ctx context.Context, partitionID string) error {
	//TODO implement me
	panic("implement me")
}

type datastoreOptions struct {
	DatabaseConnectionString string
}

var defaultAgentDatastoreOptions = datastoreOptions{}
var globalAgentDatastoreOptions []DatastoreOption

type DatastoreOption interface {
	apply(*datastoreOptions)
}

// funcAgentDatastoreOption wraps a function that modifies datastoreOptions into an
// implementation of the DatastoreOption interface.
type funcAgentDatastoreOption struct {
	f func(*datastoreOptions)
}

func (fdo *funcAgentDatastoreOption) apply(do *datastoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*datastoreOptions)) *funcAgentDatastoreOption {
	return &funcAgentDatastoreOption{
		f: f,
	}
}

func WithDatastoreConnectionString(connString string) DatastoreOption {
	return newFuncServerOption(func(o *datastoreOptions) {
		o.DatabaseConnectionString = connString
	})
}
