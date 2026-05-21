-- +goose Up
-- Hash existing plaintext legacy user tokens. Clients sending the raw token
-- continue to work transparently: the server hashes before lookup.
UPDATE users SET token = encode(sha256(token::bytea), 'hex') WHERE token IS NOT NULL;

-- +goose Down
-- Not reversible: raw tokens are not recoverable after hashing.
SELECT 'down migration not supported';
