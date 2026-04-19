package main

import (
	"embed"
	"io/fs"

	"github.com/nyambati/litmus/cmd"
	"github.com/nyambati/litmus/internal/server"
)

//go:embed ui/dist/*
var staticFiles embed.FS

func main() {
	publicFS, _ := fs.Sub(staticFiles, "ui/dist")
	server.SetStaticFS(publicFS)
	cmd.Execute()
}
