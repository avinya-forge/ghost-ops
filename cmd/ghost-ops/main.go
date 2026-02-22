package main

import (
	"ghost-ops/cmd/ghost-ops/cli"
)

var Version = "dev"

func main() {
	cli.Execute(Version)
}
