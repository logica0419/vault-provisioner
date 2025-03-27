package kube

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetSecret(ctx context.Context, name, namespace string) (*coreV1.Secret, error) {
	secret, err := clientSet.CoreV1().Secrets(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func ApplySecret(ctx context.Context, secret *coreV1.Secret) error {
	client := clientSet.CoreV1().Secrets(secret.Namespace)

	_, err := client.Get(ctx, secret.Name, metaV1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err != nil {
		_, err = client.Create(ctx, secret, metaV1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		_, err = client.Update(ctx, secret, metaV1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
