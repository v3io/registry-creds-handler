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

func (ar *Registry) Validate() error {
	if ar.SecretName == "" {
		return errors.New("Secret Name must not be empty")
	}
	if ar.RegistryUri == "" {
		return errors.New("Registry URI must not be empty")
	}
	if ar.Namespace == "" {
		ar.Logger.DebugWith("Did not receive namespace, using `default`")
		ar.Namespace = "default"
	}
	return nil
}
