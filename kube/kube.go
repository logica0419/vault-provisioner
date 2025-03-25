package kube

import (
	"log"
	"log/slog"

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
