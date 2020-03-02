-- name: CreateAuthor :one
INSERT INTO authors (name)
VALUES ($1)
RETURNING *;

-- name: DeleteAuthor :exec
DELETE FROM authors
WHERE id = $1;

-- name: CreateBook :one
INSERT INTO books (title, author_id, isbn)
VALUES ($1, $2, $3)
RETURNING *;

-- name: BooksByAuthor :many
SELECT b.*
FROM authors a
         JOIN books b on a.id = b.author_id
WHERE a.name = $1;
