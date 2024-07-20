package pgx

import (
	"context"
	"errors"
	"fmt"
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
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE agent_id = $1 AND deleted_at IS NULL ORDER BY id", agentID)
	agent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[metadatastore.Agent])
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (a *Datastore) ReadAllAgents(ctx context.Context) ([]*metadatastore.Agent, error) {
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE deleted_at IS NULL ORDER BY id")
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
	_, err := a.db.Exec(ctx, "INSERT INTO topics (topic_id, name, internal) VALUES ($1, $2)", topic.ID, topic.Name, topic.Internal)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (a *Datastore) ReadTopic(ctx context.Context, topicID string) (*metadatastore.Topic, error) {
	rows, _ := a.db.Query(ctx, "SELECT topic_id, name, internal FROM topics WHERE topic_id = $1 AND deleted_at IS NULL", topicID)
	topic, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[metadatastore.Topic])
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (a *Datastore) ReadAllTopics(ctx context.Context) ([]*metadatastore.Topic, error) {
	rows, _ := a.db.Query(ctx, "SELECT topic_id, name, internal FROM topics WHERE deleted_at IS NULL")
	topics, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[metadatastore.Topic])
	if err != nil {
		return nil, err
	}
	return topics, nil
}

func (a *Datastore) UpdateTopic(ctx context.Context, topic *metadatastore.Topic) error {
	_, err := a.db.Exec(ctx, "UPDATE topics SET name = $1, internal = $2 WHERE topic_id = $3", topic.Name, topic.Internal, topic.ID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Datastore) DeleteTopic(ctx context.Context, topicID string) error {
	tag, err := a.db.Exec(ctx, "UPDATE topics SET deleted_at = $1 WHERE topic_id = $2", time.Now(), topicID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *Datastore) CreatePartition(ctx context.Context, partition *metadatastore.Partition) error {
	_, err := a.db.Exec(ctx, "INSERT INTO partitions (partition_id, topic_id) VALUES ($1, $2)", partition.ID, partition.TopicID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (a *Datastore) ReadPartition(ctx context.Context, partitionID string) (*metadatastore.Partition, error) {
	rows, _ := a.db.Query(ctx, "SELECT partition_id, topic_id FROM partitions WHERE partition_id = $1 AND deleted_at IS NULL", partitionID)
	partition, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[metadatastore.Partition])
	if err != nil {
		return nil, err
	}
	return &partition, nil
}

func (a *Datastore) ReadAllPartitions(ctx context.Context) ([]*metadatastore.Partition, error) {
	rows, _ := a.db.Query(ctx, "SELECT partition_id, topic_id FROM partitions WHERE deleted_at IS NULL")
	partitions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[metadatastore.Partition])
	if err != nil {
		return nil, err
	}
	return partitions, nil
}

func (a *Datastore) UpdatePartition(ctx context.Context, partition *metadatastore.Partition) error {
	_, err := a.db.Exec(ctx, "UPDATE partitions SET topic_id = $1 WHERE partition_id = $2", partition.TopicID, partition.ID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Datastore) DeletePartition(ctx context.Context, partitionID string) error {
	tag, err := a.db.Exec(ctx, "UPDATE partitions SET deleted_at = $1 WHERE partition_id = $2", time.Now(), partitionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *Datastore) RoundRobinForAz(ctx context.Context, az string, liveAgents int) (int, error) {
	if liveAgents == 0 {
		return 0, fmt.Errorf("no live agents")
	}
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var index int
	err = tx.QueryRow(ctx, "SELECT index FROM agent_round_robin WHERE availability_zone = $1 LIMIT 1;", az).Scan(&index)
	if err != nil {
		_, err := tx.Exec(ctx, "INSERT INTO agent_round_robin (availability_zone, index) VALUES ($1, $2)", az, 0)
		if err != nil {
			return 0, err
		}
	}

	newIndex := index + 1
	if newIndex >= liveAgents {
		newIndex = 0
	}

	_, err = tx.Exec(ctx, "UPDATE agent_round_robin SET index = $1 WHERE availability_zone = $2", newIndex, az)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return newIndex, nil
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
