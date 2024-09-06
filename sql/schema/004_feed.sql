-- +goose Up
CREATE TABLE feed (
  id uuid DEFAULT gen_random_uuid(),
  created_at timestamp DEFAULT now() NOT NULL,
  updated_at timestamp DEFAULT now() NOT NULL,
  name text NOT NULL,
  url text NOT NULL UNIQUE,
  user_id uuid NOT NULL, 
  CONSTRAINT fk_user
    FOREIGN KEY (user_id) 
    REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed;
