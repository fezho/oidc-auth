package app

import (
	"fmt"
	"github.com/fezho/oidc-auth-service/cmd/auth-service/app/options"
	"github.com/fezho/oidc-auth-service/version"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func CommandServe() *cobra.Command {
	opts, err := options.NewOptions()
	if err != nil {
		// TODO: use another log
		log.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use:  "auth-service",
		Long: `The auth-service is a oidc auth service...`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCommand(cmd, args, opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
		Args: cobra.ExactArgs(0),
	}

	fs := cmd.Flags()
	opts.AddFlags(fs)

	return cmd
}

func runCommand(cmd *cobra.Command, args []string, opts *options.Options) error {
	if opts.PrintVersion {
		version.PrintVersionAndExit()
	}

	// TODO: implement serve func
	fmt.Println("implement here...")
	return nil
}
