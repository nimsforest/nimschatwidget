package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configPath string

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nimschatwidget",
		Short: "Embeddable nim chat widget service",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("nimschatwidget %s\n", Version)
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")

	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
