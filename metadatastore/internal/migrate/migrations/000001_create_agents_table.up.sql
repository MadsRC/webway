CREATE TABLE IF NOT EXISTS agents(
    id SERIAL NOT NULL PRIMARY KEY,
    hostname VARCHAR(100) NOT NULL,
    port INT NOT NULL CONSTRAINT CHK_port CHECK(port >=0 AND port <= 65535),
    availability_zone VARCHAR(100) NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);