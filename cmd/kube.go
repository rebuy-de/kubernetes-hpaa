package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeParams struct {
	Kubeconfig string
	Namespace  string
}

func (p *KubeParams) Bind(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&p.Kubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file. Keep empty for in-cluster configuration.")
	cmd.PersistentFlags().StringVarP(
		&p.Namespace, "namespace", "n", "",
		"Namespace to watch. Keep empty for all namespaces.")
}

func (p *KubeParams) Create() (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if p.Kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", p.Kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(config)
}
