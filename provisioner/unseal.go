package provisioner

import "github.com/hashicorp/vault-client-go"

func Unseal() {
	_, _ = vault.New()
}
