package main

import "github.com/powersjcb/sqlctest/gateways/http"

func main() {
	s := http.NewHTTPServer()
	s.Start()
}
