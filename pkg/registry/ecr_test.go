package registry

import (
	"os"
	"testing"

	"github.com/nuclio/logger"
	"github.com/nuclio/loggerus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type ECRSuite struct {
	suite.Suite
}

func (suite *ECRSuite) TestEnrichAndValidateECRParams() {
	loggerInstance := suite.createLogger()
	tests := []struct {
		name             string
		abstractRegistry *abstractRegistry
		error            bool
		withEnv          bool
	}{

		// happy
		{
			name: "sanity",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "namespace",
				Creds:      "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				Endpoints:  nil,
				Token:      Token{},
			},
			error:   false,
			withEnv: false,
		},
		{
			name: "envAccessKeys",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "namespace",
				Creds:      "{\"region\": \"region\"}",
				Endpoints:  nil,
				Token:      Token{},
			},
			error:   false,
			withEnv: true,
		},

		// bad
		{
			name: "missingAccessKeyId",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "namespace",
				Creds:      "{\"region\": \"region\", \"secretAccessKey\": \"some secret access key\"}",
				Endpoints:  nil,
				Token:      Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingRegion",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "namespace",
				Creds:      "{\"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				Endpoints:  nil,
				Token:      Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingSecretName",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "",
				Namespace:  "namespace",
				Creds:      "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				Endpoints:  nil,
				Token:      Token{},
			},
			error:   true,
			withEnv: false,
		},
		{
			name: "missingNamespace",
			abstractRegistry: &abstractRegistry{
				logger:     loggerInstance,
				SecretName: "secret",
				Namespace:  "",
				Creds:      "{\"region\": \"region\", \"accessKeyID\": \"some access key id\", \"secretAccessKey\": \"some secret access key\"}",
				Endpoints:  nil,
				Token:      Token{},
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

			e := &ECR{
				abstractRegistry: test.abstractRegistry,
			}
			err := e.EnrichAndValidateECRParams()
			if !test.error {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *ECRSuite) createLogger() logger.Logger {
	parsedLogLevel, err := logrus.ParseLevel("debug")
	suite.Require().NoError(err, "Failed to parse log level")

	loggerInstance, err := loggerus.NewJSONLoggerus("ecr.suite", parsedLogLevel, os.Stdout)
	suite.Require().NoError(err, "Can't create logger")

	return loggerInstance
}

func TestECR(t *testing.T) {
	suite.Run(t, new(ECRSuite))
}
