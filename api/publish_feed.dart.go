package api

import (
	"context"
	db "github.com/imrishuroy/legal-referral-fanout-service/db/sqlc"
	"github.com/rs/zerolog/log"
)

func publishNewsFeed(store db.Store, userID string, postID int32) error {
	// Initialize the userIDs slice with the given userID as the first element
	userIDs := []string{userID}

	// Get the list of connected user IDs
	connectedUserIDs, err := store.ListConnectedUserIDs(context.Background(), userID)
	if err != nil {
		log.Error().Err(err).Msg("cannot list connected user IDs")
		return err
	}

	// Append the connected user IDs to the userIDs slice
	userIDs = append(userIDs, connectedUserIDs...)

	// Post to the news feed for each user ID
	for _, id := range userIDs {
		arg := db.PostToNewsFeedParams{
			UserID: id,
			PostID: postID,
		}

		if err := store.PostToNewsFeed(context.Background(), arg); err != nil {
			log.Error().Err(err).Str("userID", id).Msg("cannot post to news feed")
			return err
		}
	}
	return nil
}
