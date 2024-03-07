-- +goose Up

-- ('M', 'Music')
-- ('P', 'Podcast')
CREATE TYPE content_type AS ENUM ('M', 'P');

CREATE TABLE content (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    title VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    type content_type NOT NULL,
    s3_key TEXT UNIQUE,
    CONSTRAINT UniqueContent UNIQUE (user_id, title)
);

-- +goose Down
DROP TABLE content;