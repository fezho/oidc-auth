package main

import (
	"fmt"
	"os"

	"github.com/fezho/oidc-auth-service/cmd/auth-service/app"
)

func main() {
	command := app.CommandServe()

	// 	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	// 	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
