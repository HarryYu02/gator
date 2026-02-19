-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    *,
    (SELECT name FROM users u WHERE feed_follows.user_id = u.id) AS user_name,
    (SELECT name FROM feeds f WHERE feed_follows.feed_id = f.id) AS feed_name;
