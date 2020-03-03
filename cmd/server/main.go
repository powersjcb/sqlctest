package main

import (
	"database/sql"
	"github.com/powersjcb/sqlctest/gateways/server"
	"github.com/powersjcb/sqlctest/internal/usecases"
	"log"
)

func main() {
	conn, err := sql.Open("postgres", "host=127.0.0.1 dbname=sqlctest sslmode=disable")
	if err != nil {
		log.Fatal(err.Error())
	}

	u := usecases.NewCoreUsecases(usecases.UsecasesArgs{Conn: conn})

	s := server.New(server.Args{
		Usecases: u,
	})
	s.Start()
}
