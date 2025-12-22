-- name: CreateGroup :one
INSERT INTO groups (name)
VALUES (?)
RETURNING *;

-- name: ListGroups :many
SELECT *
FROM groups
ORDER BY name ASC;

-- name: DeleteGroup :exec
delete from groups
where name = ?;

