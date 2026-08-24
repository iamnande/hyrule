-- name: Upsert :one
INSERT INTO pings (name, kind)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE
    SET kind = EXCLUDED.kind,
        last_seen_at = now()
RETURNING *;

-- name: List :many
SELECT * FROM pings ORDER BY name;
