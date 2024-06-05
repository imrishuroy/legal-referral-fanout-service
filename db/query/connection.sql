-- name: ListConnectedUserIDs :many
SELECT
    CASE
        WHEN sender_id = @user_id::text THEN recipient_id
        ELSE sender_id
        END AS user_id
FROM connections
WHERE sender_id = @user_id::text OR recipient_id = @user_id::text;


