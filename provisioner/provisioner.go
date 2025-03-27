package provisioner

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/hashicorp/vault-client-go"

	"github.com/logica0419/vault-provisioner/kube"
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

	unsealOpt UnsealOption
}

func New(ctx context.Context, opt VaultOption, unsealOpt UnsealOption) (*Provisioner, error) {
	if opt.Namespace == "" {
		namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, err
		}

		opt.Namespace = string(namespace)
	}

	pods, err := kube.GetPods(ctx, opt.Namespace)
	if err != nil {
		return nil, err
	}

	provisioner := &Provisioner{
		vaultClients: make([]*vault.Client, opt.Replicas),

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
