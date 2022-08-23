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
	case registry.ECRRegistryKind:
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
