package factory

import (
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/ecr"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

// CreateRegistry creates a registry based on a requested kind (registryKind)
func CreateRegistry(parentLogger logger.Logger,
	registryKind string,
	secretName string,
	namespace string,
	creds string,
	registryUri string) (registry.Registry, error) {

	var newRegistry registry.Registry
	var err error

	switch registryKind {
	case string(registry.ECRRegistryKind):
		newRegistry, err = ecr.NewRegistry(parentLogger,
			secretName,
			namespace,
			creds,
			registryUri)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create ECR kind")
		}
		if err := newRegistry.EnrichAndValidate(); err != nil {
			return nil, errors.Wrap(err, "Failed to enrich and validate")
		}
	default:
		return nil, errors.Errorf("Unsupported registry kind: %s", registryKind)
	}

	return newRegistry, nil
}
