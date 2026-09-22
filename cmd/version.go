package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "v1.0.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Unclick",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Unclick " + version)
		fmt.Println("https://github.com/Perruer/unclick")
	},
}
