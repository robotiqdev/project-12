package main

import (
	"log"
	"net/http"

	"github.com/robotiqdev/project-12/server"
	"github.com/robotiqdev/project-12/version"
)

func main() {
	srv := server.New(server.Config{
		Version: version.Version,
	})

	log.Fatal(http.ListenAndServe(":8080", srv))
}
