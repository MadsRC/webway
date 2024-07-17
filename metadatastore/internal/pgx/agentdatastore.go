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

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("already exists")
)

type AgentDatastore struct {
	opts agentDatastoreOptions
	db   *pgxpool.Pool
}

func NewAgentDatastore(opt ...AgentDatastoreOption) (*AgentDatastore, error) {
	opts := defaultAgentDatastoreOptions
	for _, o := range globalAgentDatastoreOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	a := &AgentDatastore{opts: opts}

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

func (a *AgentDatastore) GracefulStop() {
	a.db.Close()
}

func (a *AgentDatastore) Create(ctx context.Context, agent *metadatastore.Agent) error {
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

func (a *AgentDatastore) Read(ctx context.Context, agentID string) (*metadatastore.Agent, error) {
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE agent_id = $1 AND deleted_at IS NULL", agentID)
	agent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[metadatastore.Agent])
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (a *AgentDatastore) ReadAll(ctx context.Context) ([]*metadatastore.Agent, error) {
	rows, _ := a.db.Query(ctx, "SELECT agent_id, availability_zone, last_seen FROM agents WHERE deleted_at IS NULL")
	agents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[metadatastore.Agent])
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (a *AgentDatastore) Update(ctx context.Context, agent *metadatastore.Agent) error {
	_, err := a.db.Exec(ctx, "UPDATE agents SET availability_zone = $1, last_seen = $2, deleted_at = null WHERE agent_id = $3", agent.AvailabilityZone, time.Now(), agent.ID)
	if err != nil {
		return err
	}
	return nil
}

func (a *AgentDatastore) Delete(ctx context.Context, agentID string) error {
	tag, err := a.db.Exec(ctx, "UPDATE agents SET deleted_at = $1 WHERE agent_id = $2", time.Now(), agentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type agentDatastoreOptions struct {
	DatabaseConnectionString string
}

var defaultAgentDatastoreOptions = agentDatastoreOptions{}
var globalAgentDatastoreOptions []AgentDatastoreOption

type AgentDatastoreOption interface {
	apply(*agentDatastoreOptions)
}

// funcAgentDatastoreOption wraps a function that modifies agentDatastoreOptions into an
// implementation of the AgentDatastoreOption interface.
type funcAgentDatastoreOption struct {
	f func(*agentDatastoreOptions)
}

func (fdo *funcAgentDatastoreOption) apply(do *agentDatastoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*agentDatastoreOptions)) *funcAgentDatastoreOption {
	return &funcAgentDatastoreOption{
		f: f,
	}
}

func WithAgentDatastoreConnectionString(connString string) AgentDatastoreOption {
	return newFuncServerOption(func(o *agentDatastoreOptions) {
		o.DatabaseConnectionString = connString
	})
}
