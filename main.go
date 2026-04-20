package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/nyambati/litmus/cmd"
	"github.com/nyambati/litmus/internal/server"
)

//go:embed ui/dist/*
var staticFiles embed.FS

func main() {
	publicFS, err := fs.Sub(staticFiles, "ui/dist")
	if err != nil {
		log.Fatalf("failed to initialize UI filesystem: %v", err)
	}
	server.SetStaticFS(publicFS)
	cmd.Execute()
}
