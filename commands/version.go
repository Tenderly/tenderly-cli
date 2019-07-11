package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var CurrentCLIVersion string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current CLI version: %s\n\n"+
			"To report a bug or give feedback send us an email at support@tenderly.app or join our Discord channel at https://discord.gg/eCWjuvt\n",
			CurrentCLIVersion,
		)
	},
}

func SetCurrentCLIVersion(version string) {
	CurrentCLIVersion = version
	if !strings.HasPrefix(CurrentCLIVersion, "v") {
		CurrentCLIVersion = fmt.Sprintf("v%s", CurrentCLIVersion)
	}

	CheckVersion(false, false)
}
