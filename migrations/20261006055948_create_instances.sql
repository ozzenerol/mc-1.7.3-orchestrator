-- +goose Up
CREATE TYPE instance_status AS ENUM ('ONLINE', 'OFFLINE', 'STARTING');

CREATE TABLE instances (
  id               UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
  host_id          UUID            NOT NULL REFERENCES hosts (id) ON DELETE CASCADE,
  port             INTEGER         NOT NULL CHECK (port BETWEEN 1024 AND 65535),
  max_players      INTEGER         NOT NULL CHECK (max_players > 0),
  dedicated_ram_mb INTEGER         NOT NULL CHECK (dedicated_ram_mb > 0),
  log_path         TEXT            NOT NULL,
  created_at       TIMESTAMPTZ     NOT NULL DEFAULT now(),
  status           instance_status NOT NULL DEFAULT 'OFFLINE',
  UNIQUE (host_id, port)
);

-- +goose Down
DROP TABLE instances;
DROP TYPE instance_status;
