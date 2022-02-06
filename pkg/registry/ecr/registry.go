package ecr

import (
	"context"
	"encoding/json"
	"os"

	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/abstract"
	"github.com/v3io/registry-creds-handler/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type Registry struct {
	*abstract.Registry
	awsCreds registry.AWSCreds
}

func NewRegistry(parentLogger logger.Logger,
	secretName string,
	namespace string,
	creds string,
	registryUri string) (*Registry, error) {
	newRegistry := &Registry{}

	// create base
	abstractRegistry, err := abstract.NewRegistry(parentLogger, newRegistry, secretName, namespace, creds, registryUri)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}

	newRegistry.Registry = abstractRegistry
	return newRegistry, nil
}

func (r *Registry) EnrichAndValidate() error {

	if err := r.Enrich(); err != nil {
		return errors.Wrap(err, "Failed to enrich ECR Registry")
	}

	if err := r.Validate(); err != nil {
		return errors.Wrap(err, "Failed to validate ECR Registry")
	}

	return nil
}

func (r *Registry) Enrich() error {

	// parse aws credentials
	var awsCreds registry.AWSCreds
	if err := json.Unmarshal([]byte(r.Creds), &awsCreds); err != nil {
		return errors.Wrap(err, "Failed to parse AWS credentials")
	}

	awsCreds.Region = util.GetFirstNonEmptyString([]string{awsCreds.Region, os.Getenv("AWS_DEFAULT_REGION")})
	awsCreds.AccessKeyID = util.GetFirstNonEmptyString([]string{awsCreds.Region, os.Getenv("AWS_ACCESS_KEY_ID")})
	awsCreds.SecretAccessKey = util.GetFirstNonEmptyString([]string{awsCreds.Region, os.Getenv("AWS_SECRET_ACCESS_KEY")})
	r.awsCreds = awsCreds

	return nil
}

func (r *Registry) Validate() error {
	if err := r.Registry.ValidateParameters(); err != nil {
		return errors.Wrap(err, "Failed to validate base parameters")
	}

	if r.awsCreds.Region == "" {
		return errors.New("AWS Region is required")
	}

	if r.awsCreds.AccessKeyID == "" {
		return errors.New("AWS Access Key ID is required")
	}

	if r.awsCreds.SecretAccessKey == "" {
		return errors.New("AWS Secret Access Key is required")
	}

	return nil
}

func (r *Registry) GetAuthToken(ctx context.Context) (*registry.Token, error) {
	ecrClient := r.createECRClient(ctx)

	r.Logger.DebugWithCtx(ctx, "Getting authorization token", "SecretName", r.SecretName, "Namespace", r.Namespace)
	resp, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})

	if err != nil {
		r.Logger.WarnWithCtx(ctx, "Failed to get ECR authorization token", "error", err.Error())
		return nil, errors.Wrap(err, "Failed to get authorization token from ecr client")
	}
	r.Logger.DebugWithCtx(ctx, "Got GetAuthorizationToken response from ECR")

	// This token has access to any ECR registry that the IAM principal has access to
	for _, auth := range resp.AuthorizationData {
		token := &registry.Token{
			SecretName:  r.SecretName,
			AccessToken: *auth.AuthorizationToken,
			RegistryUri: r.RegistryUri,
		}
		r.Logger.InfoWithCtx(ctx, "Got authorization token", "ExpiresAt", auth.ExpiresAt)
		return token, nil
	}

	return nil, errors.New("Failed to retrieve ECR access token")
}

func (r *Registry) createECRClient(ctx context.Context) *ecr.ECR {
	r.Logger.DebugWithCtx(ctx, "Creating ECR Client", "region", r.awsCreds.Region, "assumeRole", r.awsCreds.AssumeRole)
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(r.awsCreds.Region),
		Credentials: credentials.NewStaticCredentials(r.awsCreds.AccessKeyID,
			r.awsCreds.SecretAccessKey,
			"")}))

	if r.awsCreds.AssumeRole != "" {
		creds := stscreds.NewCredentials(sess, r.awsCreds.AssumeRole)
		return ecr.New(sess, &aws.Config{Credentials: creds})
	}

	return ecr.New(sess)
}
