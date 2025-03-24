package kube

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientSet *kubernetes.Clientset

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panic(err)
	}

	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}
}
