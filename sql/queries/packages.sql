-- name: AddPackageWithVersion :one
INSERT INTO packages (name, url, version) 
VALUES (?, ?, ?)
ON CONFLICT (name) DO UPDATE 
SET is_deleted = false, url = excluded.url, version = excluded.version
RETURNING *;

-- name: GetIDByName :one
SELECT id FROM packages WHERE name = ?;

-- name: UpdatePackageUsage :exec
UPDATE packages 
SET freq = freq + 1, last_used = CURRENT_TIMESTAMP 
WHERE url = ?;

-- name: UpdatePackageByName :one
UPDATE packages
SET url = ?, version = ?
WHERE name = ?
RETURNING *;

-- name: GetURLsByNames :many
SELECT name, url, version 
FROM packages 
WHERE name IN (sqlc.slice('names'));

-- name: UpdatePackage :one
UPDATE packages
SET name = ?, url = ?, version = ?
WHERE id = ?
RETURNING *;

-- name: ListPackagesByLastUsed :many
SELECT * FROM packages
WHERE is_deleted = false
ORDER BY last_used DESC
LIMIT ?;

-- name: MarkDeleteByName :exec
UPDATE packages
SET is_deleted = true, updated_at = CURRENT_TIMESTAMP
WHERE name IN (sqlc.slice('names'));

-- name: MarkDeleteFalse :exec
UPDATE packages
set is_deleted = false, updated_at = CURRENT_TIMESTAMP
where name in (sqlc.slice('names'));

-- name: DeletePackagesByName :exec
DELETE FROM packages
WHERE name IN (sqlc.slice('names'));

-- name: CleanDatabase :exec
DELETE from packages
where is_deleted = true;


-- name: GetPackageIDByURL :one
SELECT id
FROM packages
WHERE url = ?;


-- name: ListPackagesByFrequency :many
SELECT * FROM packages
WHERE is_deleted = false
ORDER BY freq DESC
LIMIT ?;

-- name: GetPackageByID :one
SELECT * FROM packages WHERE id = ? and is_deleted = false;

-- name: GetPackageByName :one
SELECT * FROM packages WHERE name =? and is_deleted = false;
