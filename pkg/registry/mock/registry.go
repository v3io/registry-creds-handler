/*
Copyright 2022 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/
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
