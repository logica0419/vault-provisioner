package provisioner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hashicorp/vault-client-go"

	"github.com/logica0419/vault-provisioner/kube"
	"github.com/logica0419/vault-provisioner/storage"
)

type VaultOption struct {
	// Name of the Vault StatefulSet
	Name string `json:"name" mapstructure:"name" yaml:"name"`
	// Replicas of the Vault StatefulSet
	Replicas int `json:"replicas" mapstructure:"replicas" yaml:"replicas"`
	// Namespace of the Vault Instance
	Namespace string `json:"namespace" mapstructure:"namespace" yaml:"namespace"`
	// Port of the Vault Instance
	Port int `json:"port" mapstructure:"port" yaml:"port"`
}

type Provisioner struct {
	vaultClients []*vault.Client
	keyStorage   storage.KeyStorage

	unsealOpt UnsealOption
}

func New(
	ctx context.Context, keyStorage storage.KeyStorage,
	opt VaultOption, unsealOpt UnsealOption,
) (*Provisioner, error) {
	var err error

	opt.Namespace, err = kube.GetNamespaceIfEmpty(opt.Namespace)
	if err != nil {
		return nil, err
	}

	pods, err := kube.GetPods(ctx, opt.Namespace)
	if err != nil {
		return nil, err
	}

	provisioner := &Provisioner{
		vaultClients: make([]*vault.Client, opt.Replicas),
		keyStorage:   keyStorage,

		unsealOpt: unsealOpt,
	}

	for i := range opt.Replicas {
		podIP, err := kube.GetPodIP(fmt.Sprintf("%s-%d", opt.Name, i), pods)
		if err != nil {
			return nil, err
		}

		client, err := vault.New(
			vault.WithAddress("http://"+podIP+":8200"),
			vault.WithTLS(vault.TLSConfiguration{
				InsecureSkipVerify: true,
			}),
		)
		if err != nil {
			return nil, err
		}

		provisioner.vaultClients[i] = client
	}

	return provisioner, nil
}

func (p *Provisioner) Run(ctx context.Context) error {
	if p.unsealOpt.Enabled {
		slog.Info("Starting unseal process")

		if err := p.Unseal(ctx); err != nil {
			return err
		}

		slog.Info("Unseal process completed")
	}

	return nil
}

func (p *Provisioner) Authenticate(token string) error {
	for _, client := range p.vaultClients {
		if err := client.SetToken(token); err != nil {
			return err
		}
	}

	return nil
}
