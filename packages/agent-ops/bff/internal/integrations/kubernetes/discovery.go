package kubernetes

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// TODO: Most of the logic can be reused via other client functions in this package. For now, this is here to allow discovery service to get a basic client

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return kubernetes.NewForConfig(config)
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client from kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client from kubeconfig: %w", err)
	}
	return clientset, nil
}
