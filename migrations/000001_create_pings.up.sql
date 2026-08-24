CREATE TABLE pings (
    name text PRIMARY KEY,
    kind text NOT NULL CHECK (kind IN ('host', 'service', 'app')),
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now()
);
