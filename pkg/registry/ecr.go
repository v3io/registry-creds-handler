package registry

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type ECR struct {
	*abstractRegistry
	region          string
	assumeRole      string
	accessKeyID     string
	secretAccessKey string
}

func NewECRRegistry(parentLogger logger.Logger,
	secretName string,
	namespace string,
	endpoints []string,
	region string,
	assumeRole string,
	accessKeyID string,
	secretAccessKey string,
) (*ECR, error) {
	newECR := &ECR{
		region:          region,
		assumeRole:      assumeRole,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
	}

	abstractRegistry, err := newAbstractRegistry(parentLogger.GetChild("ecr"), secretName, namespace, endpoints)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}
	newECR.abstractRegistry = abstractRegistry
	return newECR, nil
}

func (e *ECR) ValidateECRParams() error {
	err := e.validateAbstractParams()
	if err != nil {
		return errors.Wrap(err, "Failed abstract registry params validation")
	}
	if e.region == "" {
		return errors.New("AWS Region is required")
	}
	if e.accessKeyID == "" {
		return errors.New("AWS Access Key ID is required")
	}
	if e.secretAccessKey == "" {
		return errors.New("AWS Secret Access Key is required")
	}

	return nil
}

func (e *ECR) GetAuthorizationToken() (*string, error) {

	ecrClient := e.createECRClient()
	resp, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})

	if err != nil {
		e.logger.WarnWith("Failed to get ECR authorization token", "error", err.Error())
		return nil, errors.Wrap(err, "Failed to get authorization token from ecr client")
	}

	// This token has access to any ECR registry that the IAM principal has access to
	for _, auth := range resp.AuthorizationData {
		return auth.AuthorizationToken, nil
	}

	return nil, errors.New("Failed to retrieve ECR access token")
}

func (e *ECR) createECRClient() *ecr.ECR {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(e.region),
		Credentials: credentials.NewStaticCredentials(e.accessKeyID,
			e.secretAccessKey,
			"")}))

	if e.assumeRole != "" {
		creds := stscreds.NewCredentials(sess, e.assumeRole)
		return ecr.New(sess, &aws.Config{Credentials: creds})
	}

	return ecr.New(sess)
}
