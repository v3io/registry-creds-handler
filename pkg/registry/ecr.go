package registry

import (
	"encoding/json"
	"os"

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
	awsCreds AWSCreds
}

func NewECRRegistry(parentLogger logger.Logger,
	secretName string,
	namespace string,
	creds string,
	endpoints []string,
) (*ECR, error) {

	abstractRegistry, err := newAbstractRegistry(parentLogger.GetChild("ecr"), secretName, namespace, creds, endpoints)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}

	newECR := &ECR{
		abstractRegistry: abstractRegistry,
	}
	return newECR, nil
}

func (e *ECR) EnrichAndValidateECRParams() error {
	err := e.validateAbstractParams()
	if err != nil {
		return errors.Wrap(err, "Failed abstract registry params validation")
	}

	// parse aws credentials
	var awsCreds AWSCreds
	if err := json.Unmarshal([]byte(e.Creds), &awsCreds); err != nil {
		return errors.Wrap(err, "Failed to parse AWS credentials")
	}

	if awsCreds.Region == "" {
		return errors.New("AWS Region is required")
	}

	if awsCreds.AccessKeyID == "" {
		e.logger.Info("Did not receive AWS Access Key ID, checking env")
		awsCreds.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		if awsCreds.AccessKeyID == "" {
			return errors.New("AWS Access Key ID is required")
		}
	}

	if awsCreds.SecretAccessKey == "" {
		e.logger.Info("Did not receive AWS Secret Access Key, checking env")
		awsCreds.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
		if awsCreds.SecretAccessKey == "" {
			return errors.New("AWS Secret Access Key is required")
		}
	}

	e.awsCreds = awsCreds

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
		Region: aws.String(e.awsCreds.Region),
		Credentials: credentials.NewStaticCredentials(e.awsCreds.AccessKeyID,
			e.awsCreds.SecretAccessKey,
			"")}))

	if e.awsCreds.AssumeRole != "" {
		creds := stscreds.NewCredentials(sess, e.awsCreds.AssumeRole)
		return ecr.New(sess, &aws.Config{Credentials: creds})
	}

	return ecr.New(sess)
}
