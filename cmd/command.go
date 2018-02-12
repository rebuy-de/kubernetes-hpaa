package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubernetes-hpaa",
		Short: "Tames the Horizontal Pod Autoscaler, so stops to happily kill 90% of the replicas, all of a sudden.",
		Run:   run,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
		},
	}

	cmd.AddCommand(NewVersionCommand())

	return cmd
}

func run(cmd *cobra.Command, args []string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	v1.NewHorizontalPodAutoscalerInformer(clientset, "default", 30*time.Second, cache.Indexers{})
}
