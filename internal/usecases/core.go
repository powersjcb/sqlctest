package usecases

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/powersjcb/sqlctest/internal/usecases/entities/db"
)

type Usecases interface {
	CreateAuthor(ctx context.Context, name string) (db.Author, error)
	CreateBook(ctx context.Context) (db.Book, error)
}

type CoreUsecases struct {
	q *db.Queries
}

func NewCoreUsecases() CoreUsecases {
	conn, err := sql.Open("postgres", "host=127.0.0.1 dbname=sqlctest sslmode=disable")
	if err != nil {
		panic(err)
	}
	fmt.Println("opened connection to database successfully")
	q := db.New(conn)
	return CoreUsecases{
		q: q,
	}
}

func (u *CoreUsecases) CreateAuthor(ctx context.Context, name string) (db.Author, error) {
	author, err := u.q.CreateAuthor(ctx, sql.NullString{String: name, Valid: true})
	fmt.Println("created author:", author)
	return author, err
}

func (u *CoreUsecases) CreateBook(ctx context.Context, params db.CreateBookParams) (db.Book, error) {
	book, err := u.q.CreateBook(ctx, params)
	fmt.Println("created book")
	return book, err
}
