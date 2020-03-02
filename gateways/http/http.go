package http

import (
	"context"
	"fmt"
	"github.com/powersjcb/sqlctest/internal/usecases"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type key int

const (
	requestIDKey key = 0
)

var healthy int32

type Server struct {
	u usecases.CoreUsecases
}

func NewHTTPServer() Server {
	u := usecases.NewCoreUsecases()
	return Server{u: u}
}

func (s *Server) Start() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := http.NewServeMux()
	router.Handle("/", index())
	router.Handle("/healthz", healthz())
	router.HandleFunc("/author", s.Authors)

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	httpServer := http.Server{
		Addr:     ":8888",
		Handler:  tracing(nextRequestID)(logging(logger)(router)),
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
		}
		_, err = s.u.CreateAuthor(r.Context(), string(body))
		if err != nil {
			w.Write([]byte(err.Error())) // nolint
			w.WriteHeader(500)
		}
		w.Write([]byte("testing"))
		w.WriteHeader(200)
	} else {
		w.Write([]byte("not a post request"))
	}
}
