-- name: GetAppUser :one
SELECT id, name
FROM app_user
ORDER BY id
LIMIT 1;
