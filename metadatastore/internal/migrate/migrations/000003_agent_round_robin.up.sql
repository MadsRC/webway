CREATE TABLE IF NOT EXISTS agent_round_robin(
    id SERIAL NOT NULL PRIMARY KEY,
    availability_zone VARCHAR(100) NOT NULL,
    index INT NOT NULL
);