package ecr

import (
	"os"
	"testing"

	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/abstract"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/stretchr/testify/suite"
)

type ECRSuite struct {
	suite.Suite
}

func (suite *ECRSuite) TestEnrichAndValidateECRParams() {
	loggerInstance, _ := util.CreateLogger("ecr.suite", true, os.Stdout, "humanreadable")
	tests := []struct {
		name             string
		abstractRegistry *abstract.Registry
		error            bool
		withEnv          bool
	}{

		// happy
		{
			name: "sanity",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "secret",
				Namespace:   "namespace",
				Creds:       "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   false,
			withEnv: false,
		},
		{
			name: "envAccessKeys",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "secret",
				Namespace:   "namespace",
				Creds:       "{\"region\": \"region\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   false,
			withEnv: true,
		},

		// bad
		{
			name: "missingAccessKeyId",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "secret",
				Namespace:   "namespace",
				Creds:       "{\"region\": \"region\", \"secretAccessKey\": \"some secret access key\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingRegion",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "secret",
				Namespace:   "namespace",
				Creds:       "{\"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingSecretName",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "",
				Namespace:   "namespace",
				Creds:       "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingNamespace",
			abstractRegistry: &abstract.Registry{
				Logger:      loggerInstance,
				SecretName:  "secret",
				Namespace:   "",
				Creds:       "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				RegistryUri: "",
				Token:       registry.Token{},
			},
			error:   true,
			withEnv: false,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			if test.withEnv {
				err := os.Setenv("AWS_ACCESS_KEY_ID", "some access key id")
				suite.Require().NoError(err)
				err = os.Setenv("AWS_SECRET_ACCESS_KEY", "some secret access key")
				suite.Require().NoError(err)
			}

			r := &Registry{
				Registry: test.abstractRegistry,
			}
			err := r.EnrichAndValidate()
			if !test.error {
				suite.Require().NoError(err)
			}
		})
	}
}

func TestECR(t *testing.T) {
	suite.Run(t, new(ECRSuite))
}
