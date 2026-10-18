-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

CREATE SCHEMA IF NOT EXISTS events;

CREATE TABLE IF NOT EXISTS events.locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    fuzzy_district VARCHAR(128) NOT NULL,
    coordinates GEOGRAPHY(Point, 4326) NOT NULL,
    external_source_id VARCHAR(255) UNIQUE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_locations_coordinates_gist 
ON events.locations USING GIST (coordinates);


CREATE TABLE IF NOT EXISTS events.events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES events.locations(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NULL,
    max_participants INT NULL CHECK (max_participants >= 2),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'finished')),
    is_crawled BOOLEAN NOT NULL DEFAULT TRUE,
    external_url TEXT UNIQUE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_location_id ON events.events (location_id);
CREATE INDEX IF NOT EXISTS idx_events_status_starts ON events.events (status, starts_at);


CREATE TABLE IF NOT EXISTS events.event_tags (
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    interest_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, interest_id)
);

CREATE INDEX IF NOT EXISTS idx_event_tags_interest_id ON events.event_tags (interest_id);


CREATE TABLE IF NOT EXISTS events.event_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'joined' CHECK (status IN ('joined', 'left', 'kicked')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_event_participants_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_event_participants_user_id ON events.event_participants (user_id);


CREATE TABLE IF NOT EXISTS events.pings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    message VARCHAR(255) NOT NULL,
    radius_meters INT NOT NULL DEFAULT 1000 CHECK (radius_meters BETWEEN 100 AND 10000),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pings_event_id ON events.pings (event_id);
CREATE INDEX IF NOT EXISTS idx_pings_expires_at ON events.pings (expires_at);


CREATE TABLE IF NOT EXISTS events.check_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events.events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    coordinates GEOGRAPHY(Point, 4326) NOT NULL,
    distance_meters FLOAT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_check_ins_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_check_ins_coordinates_gist 
ON events.check_ins USING GIST (coordinates);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.check_ins;
DROP TABLE IF EXISTS events.pings;
DROP TABLE IF EXISTS events.event_participants;
DROP TABLE IF EXISTS events.event_tags;
DROP TABLE IF EXISTS events.events;
DROP TABLE IF EXISTS events.locations;

DROP SCHEMA IF EXISTS events CASCADE;
-- +goose StatementEnd