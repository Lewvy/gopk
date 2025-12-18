-- name: AddPackageWithoutVersion :one
INSERT into packages (
	name, url
)
VALUES (
	?, ?
) RETURNING *;

-- name: AddPackageWithVersion :one
INSERT into packages (name, url, version) VALUES (
	?, ?, ?
	) RETURNING *;

-- name: GetIDByName :one
select id from packages where name = ?;

-- name: GetPackageURLByName :one
select url from packages where name = ?;

-- name: UpdatePackage :one
UPDATE packages
set name = ?, url = ?, version = ?
where id = ?
RETURNING *;

-- name: GetPackgeByID :one
select * from packages where id = ?;

-- name: GetPackageByName :one
select * from packages where name = ?;
