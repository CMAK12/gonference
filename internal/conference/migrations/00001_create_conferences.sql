-- +goose Up
CREATE SCHEMA IF NOT EXISTS conference;

CREATE TABLE IF NOT EXISTS conference.conferences
(
    id              TEXT PRIMARY KEY,
    name            TEXT        NOT NULL DEFAULT '',
    creator_id      TEXT        NOT NULL DEFAULT '',
    invited_members TEXT        NOT NULL DEFAULT '',
    token           TEXT        NOT NULL,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS conferences_creator_id_idx ON conference.conferences (creator_id);
CREATE INDEX IF NOT EXISTS conferences_start_time_idx ON conference.conferences (start_time);

-- +goose Down
DROP TABLE IF EXISTS conference.conferences;

DROP SCHEMA IF EXISTS conference;
