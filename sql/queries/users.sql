-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_passwords, is_chirpy_red)
VALUES (gen_random_uuid(), NOW(), NOW(), $1, $2, false) RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: DeleteAllUser :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateEmailPassword :one
UPDATE users
SET email = $2, hashed_passwords = $3, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpgradeUserByID :one
UPDATE users
SET is_chirpy_red = true
WHERE id = $1 RETURNING *;