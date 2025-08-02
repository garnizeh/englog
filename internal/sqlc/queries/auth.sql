-- EngLog Authentication Queries
-- JWT token management and session handling

-- name: CreateRefreshTokenDenylist :exec
INSERT INTO refresh_token_denylist (jti, user_id, expires_at, reason)
VALUES ($1, $2, $3, COALESCE($4, 'logout'));

-- name: IsRefreshTokenDenylisted :one
SELECT EXISTS(
    SELECT 1 FROM refresh_token_denylist
    WHERE jti = $1
);

-- name: CleanupExpiredDenylistedTokens :exec
DELETE FROM refresh_token_denylist
WHERE expires_at < NOW() - INTERVAL '7 days';

-- name: GetDenylistedTokensByUser :many
SELECT * FROM refresh_token_denylist
WHERE user_id = $1
ORDER BY denylisted_at DESC;

-- name: CreateUserSession :one
INSERT INTO user_sessions (
    user_id, session_token_hash, refresh_token_hash,
    expires_at, ip_address, user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions
WHERE id = $1 AND is_active = true;

-- name: GetUserSessionByToken :one
SELECT * FROM user_sessions
WHERE session_token_hash = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateSessionActivity :exec
UPDATE user_sessions
SET last_activity = NOW()
WHERE id = $1;

-- name: DeactivateSession :exec
UPDATE user_sessions
SET is_active = false
WHERE id = $1;

-- name: DeactivateUserSessions :exec
UPDATE user_sessions
SET is_active = false
WHERE user_id = $1;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions
WHERE expires_at < NOW();

-- name: GetActiveSessionsByUser :many
SELECT * FROM user_sessions
WHERE user_id = $1 AND is_active = true
ORDER BY last_activity DESC;

-- name: GetSessionCount :one
SELECT COUNT(*) FROM user_sessions
WHERE is_active = true;

-- name: ScheduleUserDeletion :one
INSERT INTO scheduled_deletions (
    user_id, scheduled_at, deletion_type, metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetScheduledDeletions :many
SELECT * FROM scheduled_deletions
WHERE status = 'pending'
  AND scheduled_at <= NOW()
ORDER BY scheduled_at ASC;

-- name: UpdateDeletionStatus :exec
UPDATE scheduled_deletions
SET status = $2, completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END
WHERE id = $1;

-- name: GetUserDeletionRequests :many
SELECT * FROM scheduled_deletions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CancelDeletionRequest :exec
UPDATE scheduled_deletions
SET status = 'cancelled'
WHERE id = $1 AND user_id = $2 AND status = 'pending';
