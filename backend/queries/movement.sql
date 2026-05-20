-- name: InsertMovementOrder :one
INSERT INTO movement_orders (
    hero_id,
    from_node_id,
    to_node_id,
    depart_at,
    arrive_at,
    status
) VALUES (
    $1, $2, $3, $4, $5, 'in_flight'
)
RETURNING id, hero_id, from_node_id, to_node_id, depart_at, arrive_at, status, created_at;

-- name: GetActiveMovementByHero :one
SELECT id, hero_id, from_node_id, to_node_id, depart_at, arrive_at, status, created_at
FROM movement_orders
WHERE hero_id = $1
  AND status = 'in_flight'
LIMIT 1;

-- name: MarkMovementArrived :execrows
UPDATE movement_orders
SET status = 'arrived'
WHERE id = $1
  AND status = 'in_flight';

-- name: ListInFlightMovements :many
SELECT id, hero_id, from_node_id, to_node_id, depart_at, arrive_at, status, created_at
FROM movement_orders
WHERE status = 'in_flight'
ORDER BY arrive_at;
