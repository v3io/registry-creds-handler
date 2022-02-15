package ecr

import (
	"os"
	"testing"

	"github.com/v3io/registry-creds-handler/pkg/common"
	"github.com/v3io/registry-creds-handler/pkg/registry/abstract"

	"github.com/stretchr/testify/suite"
)

type ECRSuite struct {
	suite.Suite
}

func (suite *ECRSuite) TestEnrichAndValidateECRParams() {
	loggerInstance, err := common.CreateLogger("test", true, os.Stdout, "humanreadable")
	suite.Require().NoError(err)

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
				Logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "namespace",
				Creds: []byte(`{"region": "region", "accessKeyID": "some access key id",
"secretAccessKey": "some secret access key"}`),
				RegistryUri: "mock.com",
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
				Creds:       []byte{},
				RegistryUri: "mock.com",
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
				Creds:       []byte(`{"region": "region", "secretAccessKey": "some secret access key"}`),
				RegistryUri: "mock.com",
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
				Creds:       []byte(`{"accessKeyID": "some access key id", "secretAccessKey": "some secret access key"}`),
				RegistryUri: "mock.com",
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingSecretName",
			abstractRegistry: &abstract.Registry{
				Logger:     loggerInstance,
				SecretName: "",
				Namespace:  "namespace",
				Creds: []byte(`{"region": "region", "accessKeyID": "some access key id",
"secretAccessKey": "some secret access key"}`),
				RegistryUri: "mock.com",
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingNamespace",
			abstractRegistry: &abstract.Registry{
				Logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "",
				Creds: []byte(`{"region": "region", "accessKeyID": "some access key id", 
"secretAccessKey": "some secret access key"}`),
				RegistryUri: "mock.com",
			},
			error:   false,
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
				err = os.Setenv("AWS_ROLE_ARN", "some role arn")
				suite.Require().NoError(err)
				err = os.Setenv("AWS_DEFAULT_REGION", "some region")
				suite.Require().NoError(err)
			} else {
				err := os.Setenv("AWS_ACCESS_KEY_ID", "")
				suite.Require().NoError(err)
				err = os.Setenv("AWS_SECRET_ACCESS_KEY", "")
				suite.Require().NoError(err)
				err = os.Setenv("AWS_ROLE_ARN", "")
				suite.Require().NoError(err)
				err = os.Setenv("AWS_DEFAULT_REGION", "")
				suite.Require().NoError(err)
			}

			r := &Registry{
				Registry: test.abstractRegistry,
			}
			err := r.EnrichAndValidate()
			if test.error {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			if test.withEnv {
				suite.Require().Equal(os.Getenv("AWS_ACCESS_KEY_ID"), r.awsCreds.AccessKeyID)
				suite.Require().Equal(os.Getenv("AWS_SECRET_ACCESS_KEY"), r.awsCreds.SecretAccessKey)
				suite.Require().Equal(os.Getenv("AWS_ROLE_ARN"), r.awsCreds.AssumeRole)
				suite.Require().Equal(os.Getenv("AWS_DEFAULT_REGION"), r.awsCreds.Region)
			}
		})
	}
}

func TestECR(t *testing.T) {
	suite.Run(t, new(ECRSuite))
}
