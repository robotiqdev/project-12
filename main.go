package main

import (
	"log"
	"net/http"
	"time"

	"github.com/robotiqdev/project-12/internal/health"
	"github.com/robotiqdev/project-12/server"
	"github.com/robotiqdev/project-12/version"
)

func main() {
	tracker := health.NewTracker(time.Now)
	srv := server.New(server.Config{
		Version: version.Version,
	}, tracker)

	log.Fatal(http.ListenAndServe(":8080", srv))
}
