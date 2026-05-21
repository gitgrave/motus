-- +goose Up
-- Hash existing plaintext tokens. Existing API key holders continue to work
-- transparently: the client sends the raw token, the server hashes before lookup.
UPDATE api_keys SET token = encode(sha256(token::bytea), 'hex');

-- +goose Down
-- Not reversible: raw tokens are not recoverable after hashing.
SELECT 'down migration not supported';
