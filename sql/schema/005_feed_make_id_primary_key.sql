-- +goose Up
ALTER TABLE feed
ADD CONSTRAINT feed_pkey PRIMARY KEY (id);

-- +goose Down
ALTER TABLE feed
DROP CONSTRAINT feed_pkey;
