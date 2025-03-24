package kube

import (
	"context"
	"errors"
	"fmt"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPods(ctx context.Context, namespace string) (*coreV1.PodList, error) {
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

var errPodNotFound = errors.New("pod not found")

func GetPodIP(name string, pods *coreV1.PodList) (string, error) {
	for _, pod := range pods.Items {
		if pod.GetName() != name || pod.Status.PodIP == "" {
			continue
		}

		return pod.Status.PodIP, nil
	}

	return "", fmt.Errorf("%w: %s", errPodNotFound, name)
}
