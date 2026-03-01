-- Remove duplicate token_hash rows, keep the newest by created_at (and id tie-breaker).
WITH ranked AS (
  SELECT
    id,
    token_hash,
    ROW_NUMBER() OVER (
      PARTITION BY token_hash
      ORDER BY created_at DESC, id DESC
    ) AS rn
  FROM refresh_tokens
)
DELETE FROM refresh_tokens rt
USING ranked r
WHERE rt.id = r.id
  AND r.rn > 1;

-- Enforce uniqueness moving forward.
CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_tokens_token_hash
  ON refresh_tokens(token_hash);