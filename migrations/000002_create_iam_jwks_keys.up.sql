CREATE TABLE iam_jwks_keys (
    id text PRIMARY KEY,
    algorithm text NOT NULL,
    public_key text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
