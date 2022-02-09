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
