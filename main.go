package main

import (
	"github.com/nyambati/litmus/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
