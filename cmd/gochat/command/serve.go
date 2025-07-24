package cmd

import (
	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/internal/server"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			srv, err := server.SetupServer(cfg)
			if err != nil {
				return err
			}
			err = srv.ListenAndServe()

			return err
		},
	}

	return cmd
}
