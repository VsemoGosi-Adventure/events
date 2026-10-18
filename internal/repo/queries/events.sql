-- name: GetNearbyEvents :many
SELECT 
    e.id,
    e.location_id,
    e.title,
    e.description,
    e.starts_at,
    e.ends_at,
    e.max_participants,
    e.status,
    e.external_url,
    e.created_at,
    e.updated_at,
    l.id AS loc_id,
    l.name AS loc_name,
    l.address AS loc_address,
    ST_Y(l.coordinates::geometry) AS loc_lat,
    ST_X(l.coordinates::geometry) AS loc_lng,
    ST_Distance(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography) AS distance_meters,
    COALESCE(COUNT(DISTINCT ep.id) FILTER (WHERE ep.status = 'joined'), 0)::int AS current_members
FROM events.events e
JOIN events.locations l ON e.location_id = l.id
LEFT JOIN events.event_participants ep ON e.id = ep.event_id
WHERE e.status = 'active'
  AND e.starts_at >= NOW() - INTERVAL '2 hours'
  AND ST_DWithin(l.coordinates, ST_MakePoint(@user_lng::float8, @user_lat::float8)::geography, @radius_meters::float8)
GROUP BY e.id, l.id
ORDER BY distance_meters ASC
LIMIT @limit_val::int OFFSET @offset_val::int;