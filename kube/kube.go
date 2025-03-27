package kube

import (
	"log"
	"log/slog"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientSet *kubernetes.Clientset

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Warn("Failed to get in-cluster config")

		return
	}

	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}
}

func GetNamespaceIfEmpty(namespace string) (string, error) {
	if namespace == "" {
		return namespace, nil
	}

	byteNamespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}

	return string(byteNamespace), nil
}
