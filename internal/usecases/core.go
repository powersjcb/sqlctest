package usecases

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"

	"github.com/powersjcb/sqlctest/internal/usecases/entities/db"
)

type Usecases interface {
	CreateAuthor(ctx context.Context, name string) (db.Author, error)
	CreateBook(ctx context.Context, params db.CreateBookParams) (db.Book, error)
	BooksByAuthor(ctx context.Context, id int64) ([]db.Book, error)

}

type CoreUsecases struct {
	q *db.Queries
}

type UsecasesArgs struct {
	Conn db.DBTX
}

func NewCoreUsecases(args UsecasesArgs) CoreUsecases {
	q := db.New(args.Conn)
	return CoreUsecases{
		q: q,
	}
}

func (u CoreUsecases) CreateAuthor(ctx context.Context, name string) (db.Author, error) {
	return u.q.CreateAuthor(ctx, sql.NullString{String: name, Valid: true})
}

func (u CoreUsecases) CreateBook(ctx context.Context, params db.CreateBookParams) (db.Book, error) {
	return u.q.CreateBook(ctx, params)
}

func (u CoreUsecases) BooksByAuthor(ctx context.Context, id int64) ([]db.Book, error) {
	return u.q.BooksByAuthor(ctx, id)
}
