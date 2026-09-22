package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time from the release tag.
var version = "v1.0.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Unclick",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Unclick " + version)
		fmt.Println("https://github.com/Perruer/unclick")
	},
}
