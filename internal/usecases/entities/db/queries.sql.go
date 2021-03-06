// Code generated by sqlc. DO NOT EDIT.
// source: queries.sql

package db

import (
	"context"
	"database/sql"
)

const booksByAuthor = `-- name: BooksByAuthor :many
SELECT b.id, b.title, b.author_id, b.isbn
FROM authors a
         JOIN books b on a.id = b.author_id
WHERE a.id = $1
`

func (q *Queries) BooksByAuthor(ctx context.Context, id int64) ([]Book, error) {
	rows, err := q.db.QueryContext(ctx, booksByAuthor, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Book
	for rows.Next() {
		var i Book
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.AuthorID,
			&i.ISBN,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createAuthor = `-- name: CreateAuthor :one
INSERT INTO authors (name)
VALUES ($1)
RETURNING id, name
`

func (q *Queries) CreateAuthor(ctx context.Context, name sql.NullString) (Author, error) {
	row := q.db.QueryRowContext(ctx, createAuthor, name)
	var i Author
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const createBook = `-- name: CreateBook :one
INSERT INTO books (title, author_id, isbn)
VALUES ($1, $2, $3)
RETURNING id, title, author_id, isbn
`

type CreateBookParams struct {
	Title    sql.NullString `json:"title"`
	AuthorID sql.NullInt64  `json:"author_id"`
	ISBN     sql.NullString `json:"isbn"`
}

func (q *Queries) CreateBook(ctx context.Context, arg CreateBookParams) (Book, error) {
	row := q.db.QueryRowContext(ctx, createBook, arg.Title, arg.AuthorID, arg.ISBN)
	var i Book
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.AuthorID,
		&i.ISBN,
	)
	return i, err
}

const deleteAuthor = `-- name: DeleteAuthor :exec
DELETE FROM authors
WHERE id = $1
`

func (q *Queries) DeleteAuthor(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteAuthor, id)
	return err
}
