-- name: GetEventByID :one
SELECT id, location_id, title, description, starts_at, ends_at, max_participants, status, is_crawled, created_at, updated_at
FROM events.events
WHERE id = $1 LIMIT 1;

-- name: ListActiveEventsNearby :many
SELECT e.id, e.title, e.starts_at, l.name AS location_name, l.fuzzy_district
FROM events.events e
JOIN events.locations l ON e.location_id = l.id
WHERE e.status = 'active'
  AND ST_DWithin(l.coordinates, ST_MakePoint($1, $2)::geography, $3)
ORDER BY e.starts_at ASC
LIMIT $4 OFFSET $5;