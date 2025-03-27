// Kubernetes Secret storage implementation
package secret

import (
	"context"
	"encoding/base64"

	"github.com/bytedance/sonic"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/logica0419/vault-provisioner/kube"
	"github.com/logica0419/vault-provisioner/storage"
)

type Option struct {
	// Kubernetes Secret name
	Name string `json:"name" mapstructure:"name" yaml:"name"`
	// Kubernetes Secret namespace
	Namespace string `json:"namespace" mapstructure:"namespace" yaml:"namespace"`
}

type Storage struct {
	name      string
	namespace string
}

var _ storage.KeyStorage = (*Storage)(nil)

func NewStorage(opt Option) (*Storage, error) {
	var err error

	opt.Namespace, err = kube.GetNamespaceIfEmpty(opt.Namespace)
	if err != nil {
		return nil, err
	}

	return &Storage{
		name:      opt.Name,
		namespace: opt.Namespace,
	}, nil
}

const (
	rootTokenKey = "root_token"
	keysKey      = "keys"
)

func (s *Storage) Get(ctx context.Context) (string, []string, error) {
	secret, err := kube.GetSecret(ctx, s.name, s.namespace)
	if err != nil {
		return "", nil, err
	}

	dst, err := base64.StdEncoding.DecodeString(string(secret.Data[rootTokenKey]))
	if err != nil {
		return "", nil, err
	}

	rootToken := string(dst)

	dst, err = base64.StdEncoding.DecodeString(string(secret.Data[keysKey]))
	if err != nil {
		return "", nil, err
	}

	keys := make([]string, 0)

	err = sonic.Unmarshal(dst, &keys)
	if err != nil {
		return "", nil, err
	}

	return rootToken, keys, nil
}

func (s *Storage) Store(ctx context.Context, rootToken string, keys []string) error {
	keysByte, err := sonic.Marshal(keys)
	if err != nil {
		return err
	}

	secret := &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      s.name,
			Namespace: s.namespace,
		},
		Data: map[string][]byte{
			rootTokenKey: []byte(base64.StdEncoding.EncodeToString([]byte(rootToken))),
			keysKey:      []byte(base64.StdEncoding.EncodeToString(keysByte)),
		},
	}

	return kube.ApplySecret(ctx, secret)
}
