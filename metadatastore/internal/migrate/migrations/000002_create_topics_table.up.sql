CREATE TABLE IF NOT EXISTS topics(
    id SERIAL NOT NULL PRIMARY KEY,
    topic_id uuid UNIQUE DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    internal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS partitions(
    id SERIAL NOT NULL PRIMARY KEY,
    topic_id uuid NOT NULL REFERENCES topics(id),
    partition_id INTEGER UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);