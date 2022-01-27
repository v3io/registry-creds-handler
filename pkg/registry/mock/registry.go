package mock

import (
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/abstract"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/stretchr/testify/mock"
)

type Registry struct {
	mock.Mock
	*abstract.Registry
}

func NewRegistry(parentLogger logger.Logger,
	secretName string,
	namespace string,
	creds string,
	registryUri string) (*Registry, error) {
	newRegistry := &Registry{}

	// create base
	abstractRegistry, err := abstract.NewRegistry(parentLogger, newRegistry, secretName, namespace, creds, registryUri)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}

	newRegistry.Registry = abstractRegistry
	return newRegistry, nil
}

func (r *Registry) EnrichAndValidate() error {
	err := r.ValidateAbstractParams()
	if err != nil {
		return errors.Wrap(err, "Failed abstract registry params validation")
	}
	return nil
}

func (r *Registry) GetAuthToken() (*registry.Token, error) {
	mockedToken := &registry.Token{
		SecretName:  r.SecretName,
		AccessToken: "mocked access token",
		RegistryUri: r.RegistryUri,
	}
	r.On("GetAuthToken").Return(mockedToken, nil).Once()
	r.Called()
	return mockedToken, nil
}
