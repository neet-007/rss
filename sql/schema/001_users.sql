-- +goose Up
CREATE TABLE users (
  id uuid DEFAULT gen_random_uuid(),
  created_at timestamp DEFAULT now() NOT NULL,
  updated_at timestamp DEFAULT now() NOT NULL,
  name text NOT NULL
);

-- +goose Down
DROP TABLE users;
