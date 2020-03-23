package main

import (
	"os"

	"github.com/fezho/oidc-auth-service/cmd/auth-service/app"
)

func main() {
	command := app.CommandServe()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

// TODO: rename to oidc-auth
