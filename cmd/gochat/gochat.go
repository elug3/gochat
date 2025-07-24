package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	cmd "github.com/elug3/gochat/cmd/gochat/command"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "gochat",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}

		log.Logger = zerolog.New(output).With().Timestamp().Logger()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmd.NewServeCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
