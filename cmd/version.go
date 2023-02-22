package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/_internal/version"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the current build version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Debug("fetching version")
			fmt.Println(version.GetStr())
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
