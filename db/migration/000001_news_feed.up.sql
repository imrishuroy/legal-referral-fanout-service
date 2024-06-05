CREATE TABLE news_feed (
    feed_id SERIAL PRIMARY KEY,
    user_id VARCHAR NOT NULL,
    post_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id)
);