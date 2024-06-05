-- name: PostToNewsFeed :exec
INSERT INTO news_feed (
    user_id,
    post_id
) VALUES (
    $1, $2
);
