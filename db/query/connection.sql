-- name: ListConnectedUserIDs :many
SELECT
    recipient_id AS user_id
FROM connections
WHERE sender_id = @user_id::text

UNION

SELECT
    sender_id AS user_id
FROM connections
WHERE recipient_id = @user_id::text;



