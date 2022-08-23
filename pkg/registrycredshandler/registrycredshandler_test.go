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
package registrycredshandler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/v3io/registry-creds-handler/pkg/common"
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/mock"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/fake"
)

type HandlerSuite struct {
	suite.Suite
}

func (suite *HandlerSuite) TestCreateOrUpdateSecretSanity() {
	loggerInstance, _ := common.CreateLogger("test", true, os.Stdout, "humanreadable")
	mockedRegistry, _ := mock.NewRegistry(loggerInstance, "secret name", "some namespace", "", "")
	mockedKubeClientSet := fake.NewSimpleClientset()
	handler, err := NewHandler(loggerInstance, mockedKubeClientSet, mockedRegistry, 0, "mock")
	suite.Require().NoError(err)

	mockedRegistry.On("GetAuthToken").Return(&registry.Token{}, nil).Once()
	err = handler.createOrUpdateSecret(context.Background())
	suite.Require().NoError(err)
}

func (suite *HandlerSuite) TestRefreshingSecretSanity() {
	loggerInstance, _ := common.CreateLogger("test", true, os.Stdout, "humanreadable")
	mockedRegistry, _ := mock.NewRegistry(loggerInstance, "secret name", "some namespace", "", "")
	mockedKubeClientSet := fake.NewSimpleClientset()
	handler, err := NewHandler(loggerInstance, mockedKubeClientSet, mockedRegistry, 10, "mock")
	suite.Require().NoError(err)

	// setup mock for called assertion
	handler.refreshRate = time.Duration(300) * time.Millisecond
	mockedToken := &registry.Token{
		SecretName:  mockedRegistry.SecretName,
		Namespace:   mockedRegistry.Namespace,
		Auth:        "username:password",
		RegistryUri: mockedRegistry.RegistryUri,
	}
	mockedRegistry.On("GetAuthToken").Return(mockedToken, nil)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err = handler.keepRefreshingSecret(ctx)
	}()

	// let the refresher start
	time.Sleep(time.Duration(1) * time.Second)

	// cancel ctx signals the refresher to stop
	cancel()

	// let the refresher stop
	time.Sleep(time.Duration(1) * time.Second)
	mockedRegistry.AssertCalled(suite.T(), "GetAuthToken")
	suite.Require().Error(err)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
