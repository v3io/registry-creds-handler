package abstract

import (
	"github.com/v3io/registry-creds-handler/pkg/registry"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type Registry struct {
	Logger      logger.Logger
	registry    registry.Registry
	SecretName  string
	Namespace   string
	Creds       string
	RegistryUri string
	Token       registry.Token
}

func NewRegistry(loggerInstance logger.Logger,
	registry registry.Registry,
	secretName string,
	namespace string,
	creds string,
	registryUri string) (*Registry, error) {
	abstractRegistry := &Registry{
		Logger:      loggerInstance.GetChild("registry"),
		registry:    registry,
		SecretName:  secretName,
		Namespace:   namespace,
		Creds:       creds,
		RegistryUri: registryUri,
	}
	return abstractRegistry, nil
}

func (ar *Registry) ValidateAbstractParams() error {
	if ar.SecretName == "" {
		return errors.New("Token Secret Name is required")
	}
	if ar.Namespace == "" {
		return errors.New("Namespace is required")
	}
	return nil
}
