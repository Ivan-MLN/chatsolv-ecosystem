-- name: CreateUser :one
INSERT INTO users(id,name,email,password_hash,platform_role) VALUES($1,$2,$3,$4,$5) RETURNING id,name,email,password_hash,platform_role,created_at,updated_at;
-- name: GetUserByEmail :one
SELECT id,name,email,password_hash,platform_role,created_at,updated_at FROM users WHERE email=$1;
-- name: GetUserByID :one
SELECT id,name,email,password_hash,platform_role,created_at,updated_at FROM users WHERE id=$1;
-- name: UpdatePassword :exec
UPDATE users SET password_hash=$2,updated_at=now() WHERE id=$1;
