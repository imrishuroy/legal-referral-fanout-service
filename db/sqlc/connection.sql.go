// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: connection.sql

package db

import (
	"context"
)

const listConnectedUserIDs = `-- name: ListConnectedUserIDs :many
SELECT
    CASE
        WHEN sender_id = $1::text THEN recipient_id
        ELSE sender_id
        END AS user_id
FROM connections
WHERE sender_id = $1::text OR recipient_id = $1::text
`

func (q *Queries) ListConnectedUserIDs(ctx context.Context, userID string) ([]interface{}, error) {
	rows, err := q.db.Query(ctx, listConnectedUserIDs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []interface{}{}
	for rows.Next() {
		var user_id interface{}
		if err := rows.Scan(&user_id); err != nil {
			return nil, err
		}
		items = append(items, user_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listConnectedUserIDs2 = `-- name: ListConnectedUserIDs2 :many
SELECT
    recipient_id AS user_id
FROM connections
WHERE sender_id = $1::text

UNION

SELECT
    sender_id AS user_id
FROM connections
WHERE recipient_id = $1::text
`

func (q *Queries) ListConnectedUserIDs2(ctx context.Context, userID string) ([]string, error) {
	rows, err := q.db.Query(ctx, listConnectedUserIDs2, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var user_id string
		if err := rows.Scan(&user_id); err != nil {
			return nil, err
		}
		items = append(items, user_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
