package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/powersjcb/sqlctest/internal/usecases"
	"github.com/powersjcb/sqlctest/internal/usecases/entities/db"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type key int

const (
	requestIDKey key = 0
)

var healthy int32

type Server struct {
	u usecases.Usecases
}

type Args struct {
	Usecases usecases.Usecases
}

func New(args Args) Server {
	return Server{u: args.Usecases}
}

func (s *Server) Start() {
	logger := log.New(os.Stdout, "server: ", log.LstdFlags)
	logger.Println("Server is starting...")

	r := mux.NewRouter()
	r.Handle("/", index())
	r.Handle("/healthz", healthz())
	r.HandleFunc("/author", s.Authors)
	r.HandleFunc("/author/{id:[0-9]+}/books", s.BooksByAuthor)
	r.HandleFunc("/books", s.Books)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	httpServer := http.Server{
		Addr:     ":8888",
		Handler:  tracing(nextRequestID)(logging(logger)(r)),
		ErrorLog: logger,
	}
	log.Fatal(httpServer.ListenAndServe().Error())
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func (s *Server) Authors(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(err.Error())) // nolint
			w.WriteHeader(500)
			return
		}
		a, err := s.u.CreateAuthor(r.Context(), string(body))
		if err != nil {
			w.Write([]byte(err.Error())) // nolint
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(fmt.Sprint(a.ID)))
		w.WriteHeader(201)
		return
	} else {
		w.Write([]byte("not a post request"))
	}
}

type book struct {
	Title    string `json:"title"`
	ISBN     string `json:"isbn"`
	AuthorID *int64 `json:"author_id"`
}

func bookToBook(b db.Book) book {
	var authorID *int64
	if b.AuthorID.Valid {
		authorID = &b.AuthorID.Int64
	}
	return book{
		Title: b.Title.String,
		ISBN: b.ISBN.String,
		AuthorID: authorID,
	}
}

func booksToBooks(books []db.Book) []book {
	res := make([]book, len(books))
	for i, _ := range books {
		res[i] = bookToBook(books[i])
	}
	return res
}

func ToSQLNullInt64(p *int64) sql.NullInt64 {
	fmt.Println(p)
	if p == nil || *p == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func (s *Server) Books(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		var b book
		err := decoder.Decode(&b)
		if err != nil {
			w.Write([]byte("bad json"))
			w.WriteHeader(500)
			return
		}

		newBook, err := s.u.CreateBook(r.Context(), db.CreateBookParams{
			Title:    sql.NullString{String: b.Title, Valid: true},
			AuthorID: ToSQLNullInt64(b.AuthorID),
			ISBN:     sql.NullString{String: b.ISBN, Valid: true},
		})
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(500)
			return
		}
		res, err := json.Marshal(bookToBook(newBook))
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(500)
			return
		}
		w.Write(res)
		w.WriteHeader(201)
	}
}

func (s *Server) BooksByAuthor(w http.ResponseWriter, r *http.Request) {
	idString := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
		return
	}
	books, err := s.u.BooksByAuthor(r.Context(), int64(id))
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
		return
	}

	res, err := json.Marshal(booksToBooks(books))
	w.Write(res)
	w.WriteHeader(200)
}
