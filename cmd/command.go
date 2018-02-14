package cmd

import (
	"time"

	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"

	"github.com/rebuy-de/rebuy-go-sdk/cmdutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/client-go/informers/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func NewRootCommand() *cobra.Command {
	app := new(App)

	cmd := &cobra.Command{
		Use:   "kubernetes-hpaa",
		Short: "Tames the Horizontal Pod Autoscaler, so stops to happily kill 90% of the replicas, all of a sudden.",
		Run:   app.Run,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
			if app.Params.GraylogAddress != "" {
				labels := map[string]interface{}{
					"facility":   "kubernetes-hpaa",
					"version":    BuildVersion,
					"commit-sha": BuildHash,
				}
				hook := graylog.NewGraylogHook(app.Params.GraylogAddress, labels)
				log.AddHook(hook)
			}
		},
	}

	app.Bind(cmd)

	cmd.AddCommand(NewVersionCommand())

	return cmd
}

type App struct {
	Params struct {
		Kubeconfig     string
		Namespace      string
		GraylogAddress string
	}
}

func (app *App) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&app.Params.Kubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file. Keep empty for in-cluster configuration.")
	cmd.PersistentFlags().StringVarP(
		&app.Params.Namespace, "namespace", "n", "",
		"Namespace to watch. Keep empty for all namespaces.")
	cmd.PersistentFlags().StringVar(
		&app.Params.GraylogAddress, "graylog-address", "",
		`Address to Graylog for logging (format: "ip:port").`)
}

func (app *App) Kubernetes() kubernetes.Interface {
	var config *rest.Config
	var err error

	if app.Params.Kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
			cmdutil.Exit(1)
			return nil
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", app.Params.Kubeconfig)
		if err != nil {
			log.Fatal(err)
			cmdutil.Exit(1)
			return nil
		}
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
		cmdutil.Exit(1)
		return nil
	}

	return kube
}

func (app *App) Run(cmd *cobra.Command, args []string) {
	log.WithFields(log.Fields{
		"Version": BuildVersion,
		"Date":    BuildDate,
		"Commit":  BuildHash,
	}).Info("kubernetes-hpaa started")
	defer log.Info("stopping kubernetes-hpaa")

	client := app.Kubernetes()
	stopCh := make(<-chan struct{})

	handler := &Handler{
		Client: client,
	}

	informer := v1.NewHorizontalPodAutoscalerInformer(
		client, app.Params.Namespace,
		30*time.Second, cache.Indexers{})
	informer.AddEventHandler(handler)
	informer.Run(stopCh)
}
