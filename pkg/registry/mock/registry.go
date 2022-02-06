package mock

import (
	"context"

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
	args := r.Called()
	return args.Error(0)
}

func (r *Registry) GetAuthToken(ctx context.Context) (*registry.Token, error) {
	args := r.Called()
	return args.Get(0).(*registry.Token), args.Error(1)
}
