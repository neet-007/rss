-- +goose Up
CREATE TABLE feed_follow (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamp DEFAULT now() NOT NULL,
  updated_at timestamp DEFAULT now() NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  feed_id UUID NOT NULL REFERENCES feed(id) ON DELETE CASCADE,
  UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follow;
