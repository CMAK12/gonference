-- +goose Up
CREATE SCHEMA IF NOT EXISTS conference;

CREATE TABLE IF NOT EXISTS conference.conferences
(
    id              TEXT PRIMARY KEY DEFAULT uuidv7(),
    name            TEXT        NOT NULL,
    creator_id      TEXT        NOT NULL,
    invited_members TEXT,
    token           TEXT        NOT NULL,
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS conferences_creator_id_idx ON conference.conferences (creator_id);
CREATE INDEX IF NOT EXISTS conferences_start_time_idx ON conference.conferences (start_time);

-- +goose Down
DROP TABLE IF EXISTS conference.conferences;

DROP SCHEMA IF EXISTS conference;
