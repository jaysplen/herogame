-- name: GetEdge :one
SELECT id, from_node_id, to_node_id, distance_units
FROM map_edges
WHERE from_node_id = $1
  AND to_node_id = $2;

-- name: ListNodes :many
SELECT id, name, x, y, kind
FROM map_nodes
ORDER BY id;

-- name: ListEdgesByNode :many
SELECT id, from_node_id, to_node_id, distance_units
FROM map_edges
WHERE from_node_id = $1
   OR to_node_id = $1
ORDER BY id;
