package main

import (
	"os"

	"github.com/ruandada/aws-lambda-boilerplate/aws-lambda-golang-docker-starter/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
