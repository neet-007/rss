-- +goose Up
ALTER TABLE users
ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- +goose Down
ALTER TABLE users
DROP CONSTRAINT users_pkey;
