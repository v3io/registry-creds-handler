package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type ECRClient struct {
	*abstractRegistry
	tokens []ECRToken
}

func NewClient(parentLogger logger.Logger) (*ECRClient, error) {
	newClient := &ECRClient{
		tokens: []ECRToken{},
	}
	abstractRegistry, err := newAbstractRegistry(parentLogger.GetChild("ecr"))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}
	newClient.abstractRegistry = abstractRegistry
	return newClient, nil
}

func (c *ECRClient) CreateECRClient(awsRegion string,
	awsAssumeRole string,
	awsAccessKeyID string,
	awsSecretAccessKey string) *ecr.ECR {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKeyID,
			awsSecretAccessKey,
			"")}))

	if awsAssumeRole != "" {
		creds := stscreds.NewCredentials(sess, awsAssumeRole)
		return ecr.New(sess, &aws.Config{Credentials: creds})
	}

	return ecr.New(sess)
}

func (c *ECRClient) GetAuthorizationToken(ecrClient *ecr.ECR) (*string, error) {

	resp, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})

	if err != nil {
		c.logger.WarnWith("Failed to get ECR authorization token", "error", err.Error())
		return nil, errors.Wrap(err, "Failed to get ECR authorization token")
	}

	// This token has access to any ECR registry that the IAM principal has access to
	for _, auth := range resp.AuthorizationData {
		return auth.AuthorizationToken, nil
	}

	return nil, errors.New("Failed to retrieve ECR access token")
}

func (c *ECRClient) AddToken(token Token, ecrClient *ecr.ECR) {
	c.tokens = append(c.tokens, ECRToken{
		Token:     token,
		ECRClient: ecrClient,
	})
}
