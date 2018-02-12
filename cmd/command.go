package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes-hpaa",
		Short: "Tames the Horizontal Pod Autoscaler, so stops to happily kill 90% of the replicas, all of a sudden.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
		},
	}

	cmd.AddCommand(NewVersionCommand())

	return cmd
}
