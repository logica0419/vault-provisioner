package kube

import "fmt"

func GetStatefulSetPodURL(name string, i int, serviceName, namespace string, port int) string {
	return fmt.Sprintf("http://%s-%d.%s.%s.svc.cluster.local:%d", name, i, serviceName, namespace, port)
}
