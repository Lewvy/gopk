-- name: AddPackageWithVersion :one
INSERT into packages (name, url, version) VALUES (
	?, ?, ?
	) RETURNING *;

-- name: GetIDByName :one
select id from packages where name = ?;

-- name: UpdatePackageUsage :exec
UPDATE packages 
SET freq = freq + 1, last_used = CURRENT_TIMESTAMP 
WHERE url = ?;

-- name: UpdatePackageByName :one
UPDATE packages
set url = ?, version = ?
where name = ?
RETURNING *;

-- name: GetURLsByNames :many
SELECT name, url, version 
FROM packages 
WHERE name IN (sqlc.slice('names'));

-- name: UpdatePackage :one
UPDATE packages
set name = ?, url = ?, version = ?
where id = ?
RETURNING *;

-- name: ListPackagesByLastUsed :many
SELECT * FROM packages
ORDER BY last_used DESC
LIMIT ?;

-- name: ListPackagesByFrequency :many
SELECT * FROM packages
ORDER BY freq DESC
LIMIT ?;

-- name: GetPackageByID :one
select * from packages where id = ?;

-- name: GetPackageByName :one
select * from packages where name = ?;

