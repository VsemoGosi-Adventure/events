-- name: GetNearbyEvents :many
SELECT 
    e.id, e.location_id, e.title, e.description, e.starts_at, e.ends_at, e.max_participants, e.status, e.is_crawled, e.external_url, e.created_at, e.updated_at,
    l.id AS loc_id, l.name AS loc_name, l.address AS loc_address, 
    ST_Y(l.coordinates::geometry)::float8 AS loc_lat,
    ST_X(l.coordinates::geometry)::float8 AS loc_lng,
    ST_Distance(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography)::float8 AS distance_meters,
    COALESCE(COUNT(DISTINCT ep.id) FILTER (WHERE ep.status = 'joined'), 0)::int AS current_members
FROM events.events e
JOIN events.locations l ON e.location_id = l.id
LEFT JOIN events.event_participants ep ON e.id = ep.event_id AND ep.status = 'joined'
WHERE e.status = 'active'
  AND e.starts_at >= NOW() - INTERVAL '2 hours'
  AND ST_DWithin(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography, @radius_meters::float8)
GROUP BY e.id, l.id
ORDER BY distance_meters ASC
LIMIT @limit_val::int OFFSET @offset_val::int;


-- name: GetEventByID :one
SELECT 
    e.id, e.location_id, e.title, e.description, e.starts_at, e.ends_at, e.max_participants, e.status, e.is_crawled, e.external_url, e.created_at, e.updated_at,
    l.id AS loc_id, l.name AS loc_name, l.address AS loc_address, 
    ST_Y(l.coordinates::geometry)::float8 AS loc_lat,
    ST_X(l.coordinates::geometry)::float8 AS loc_lng,
    COALESCE(COUNT(DISTINCT ep.id) FILTER (WHERE ep.status = 'joined'), 0)::int AS current_members
FROM events.events e
JOIN events.locations l ON e.location_id = l.id
LEFT JOIN events.event_participants ep ON e.id = ep.event_id AND ep.status = 'joined'
WHERE e.id = @event_id::uuid
GROUP BY e.id, l.id;


-- name: GetMyEvents :many
SELECT 
    e.id, e.location_id, e.title, e.description, e.starts_at, e.ends_at, e.max_participants, e.status, e.is_crawled, e.external_url, e.created_at, e.updated_at,
    l.id AS loc_id, l.name AS loc_name, l.address AS loc_address, 
    ST_Y(l.coordinates::geometry)::float8 AS loc_lat,
    ST_X(l.coordinates::geometry)::float8 AS loc_lng,
    COALESCE(COUNT(DISTINCT ep_all.id) FILTER (WHERE ep_all.status = 'joined'), 0)::int AS current_members
FROM events.event_participants ep
JOIN events.events e ON ep.event_id = e.id
JOIN events.locations l ON e.location_id = l.id
LEFT JOIN events.event_participants ep_all ON e.id = ep_all.event_id AND ep_all.status = 'joined'
WHERE ep.user_id = @user_id::uuid 
  AND ep.status = 'joined'
  AND e.status = 'active'
GROUP BY e.id, l.id
ORDER BY e.starts_at ASC
LIMIT @limit_val::int OFFSET @offset_val::int;


-- name: UpsertParticipant :exec
INSERT INTO events.event_participants (event_id, user_id, status, joined_at, updated_at)
VALUES (@event_id::uuid, @user_id::uuid, 'joined', NOW(), NOW())
ON CONFLICT (event_id, user_id) 
DO UPDATE SET status = 'joined', updated_at = NOW();


-- name: LeaveParticipant :exec
UPDATE events.event_participants
SET status = 'left', updated_at = NOW()
WHERE event_id = @event_id::uuid AND user_id = @user_id::uuid;


-- name: GetParticipantStatus :one
SELECT status FROM events.event_participants
WHERE event_id = @event_id::uuid AND user_id = @user_id::uuid;


-- name: ListEventParticipants :many
SELECT id, event_id, user_id, status, joined_at, updated_at
FROM events.event_participants
WHERE event_id = @event_id::uuid AND status = 'joined';


-- name: CreatePing :one
INSERT INTO events.pings (id, event_id, author_id, message, radius_meters, expires_at, created_at)
VALUES (gen_random_uuid(), @event_id::uuid, @author_id::uuid, @message::text, @radius_meters::int, @expires_at::timestamptz, NOW())
RETURNING id, event_id, author_id, message, radius_meters, expires_at, created_at;


-- name: GetNearbyPings :many
SELECT 
    p.id, p.event_id, p.author_id, p.message, p.radius_meters, p.expires_at, p.created_at,
    e.title AS event_title,
    ST_Y(l.coordinates::geometry)::float8 AS lat,
    ST_X(l.coordinates::geometry)::float8 AS lng,
    ST_Distance(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography)::float8 AS distance_meters
FROM events.pings p
JOIN events.events e ON p.event_id = e.id
JOIN events.locations l ON e.location_id = l.id
WHERE p.expires_at > NOW()
  AND ST_DWithin(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography, p.radius_meters::float8)
ORDER BY p.created_at DESC
LIMIT @limit_val::int OFFSET @offset_val::int;