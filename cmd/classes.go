package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"lutonite.dev/gaps-cli/gaps"
)

var (
	classesCmd = &cobra.Command{
		Use:   "classes",
		Short: "Print the current class list",
		Run: func(cmd *cobra.Command, args []string) {
			log.Debug("fetching classes")
			cfg := buildTokenClientConfiguration()
			classes := gaps.GetAllClasses(cfg, currentAcademicYear())
			fmt.Println("Classes:", classes)
		},
	}
)

func init() {
	rootCmd.AddCommand(classesCmd)
}
