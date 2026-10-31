-- +goose Up
ALTER TABLE events.locations
    ALTER COLUMN coordinates DROP NOT NULL,
    ALTER COLUMN fuzzy_district SET DEFAULT '';

ALTER TABLE events.events
    ADD COLUMN external_source VARCHAR(255),
    ADD COLUMN external_source_id VARCHAR(255),
    ADD COLUMN join_policy TEXT,
    ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

CREATE UNIQUE INDEX events_external_source_id_idx
    ON events.events (external_source, external_source_id);

-- +goose Down
DROP INDEX events.events_external_source_id_idx;
ALTER TABLE events.events
    DROP COLUMN tags,
    DROP COLUMN join_policy,
    DROP COLUMN external_source_id,
    DROP COLUMN external_source;
ALTER TABLE events.locations
    ALTER COLUMN fuzzy_district DROP DEFAULT;
