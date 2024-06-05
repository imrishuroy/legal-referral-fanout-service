// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: news_feed.sql

package db

import (
	"context"
)

const postToNewsFeed = `-- name: PostToNewsFeed :exec
INSERT INTO news_feed (
    user_id,
    post_id
) VALUES (
    $1, $2
)
`

type PostToNewsFeedParams struct {
	UserID string `json:"user_id"`
	PostID int32  `json:"post_id"`
}

func (q *Queries) PostToNewsFeed(ctx context.Context, arg PostToNewsFeedParams) error {
	_, err := q.db.Exec(ctx, postToNewsFeed, arg.UserID, arg.PostID)
	return err
}