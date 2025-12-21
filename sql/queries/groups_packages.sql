-- name: ListPackagesByGroup :many
SELECT p.*
FROM packages p
JOIN group_packages gp ON gp.package_id = p.id
JOIN groups g ON g.id = gp.group_id
WHERE g.name = ?
ORDER BY p.name ASC;

-- name: AssignPackageToGroup :exec
INSERT OR IGNORE INTO group_packages (group_id, package_id)
VALUES (?, ?);

-- name: GetPackageIDByURL :one
SELECT id
FROM packages
WHERE url = ?;


-- name: GetGroupIDByName :one
SELECT id FROM groups WHERE name = ?;


-- name: RemovePackagesFromGroup :exec
DELETE FROM group_packages
WHERE group_id = ?
AND package_id IN (sqlc.slice('package_ids'));


-- name: GetPackageIDsByURLs :many
SELECT id FROM packages
WHERE url IN (sqlc.slice('urls'));
