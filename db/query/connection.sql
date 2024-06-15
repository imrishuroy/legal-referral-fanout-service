-- name: ListConnectedUserIDs :many
SELECT
    CASE
        WHEN sender_id = @user_id::text THEN recipient_id
        ELSE sender_id
        END AS user_id
FROM connections
WHERE sender_id = @user_id::text OR recipient_id = @user_id::text;

-- name: ListConnectedUserIDs2 :many
SELECT
    recipient_id AS user_id
FROM connections
WHERE sender_id = @user_id::text

UNION

SELECT
    sender_id AS user_id
FROM connections
WHERE recipient_id = @user_id::text;



