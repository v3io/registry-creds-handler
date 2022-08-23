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
package ecr

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/v3io/registry-creds-handler/pkg/common"
	"github.com/v3io/registry-creds-handler/pkg/registry"
	"github.com/v3io/registry-creds-handler/pkg/registry/abstract"

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
	abstractRegistry, err := abstract.NewRegistry(parentLogger.GetChild("ecr"),
		newRegistry,
		secretName,
		namespace,
		creds,
		registryUri)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create abstract registry")
	}

	newRegistry.Registry = abstractRegistry
	return newRegistry, nil
}

func (r *Registry) EnrichAndValidate() error {

	if err := r.enrich(); err != nil {
		return errors.Wrap(err, "Failed to enrich Registry")
	}

	if err := r.validate(); err != nil {
		return errors.Wrap(err, "Failed to validate Registry")
	}

	return nil
}

func (r *Registry) enrich() error {

	// parse aws credentials
	var awsCreds registry.AWSCreds
	if err := json.Unmarshal([]byte(r.Creds), &awsCreds); err != nil {
		r.Logger.WarnWith("Failed to parse json AWS credentials, checking env", "err", err.Error())
	}

	awsCreds.Region = common.GetFirstNonEmptyString(
		[]string{awsCreds.Region, strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))})
	awsCreds.AccessKeyID = common.GetFirstNonEmptyString(
		[]string{awsCreds.AccessKeyID, strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))})
	awsCreds.SecretAccessKey = common.GetFirstNonEmptyString(
		[]string{awsCreds.SecretAccessKey, strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))})
	awsCreds.AssumeRole = common.GetFirstNonEmptyString(
		[]string{awsCreds.AssumeRole, strings.TrimSpace(os.Getenv("AWS_ROLE_ARN"))})
	r.awsCreds = awsCreds

	return nil
}

func (r *Registry) validate() error {
	if err := r.Registry.Validate(); err != nil {
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

	r.Logger.DebugWithCtx(ctx, "Getting authorization token",
		"SecretName", r.SecretName,
		"Namespace", r.Namespace)
	resp, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})

	if err != nil {
		r.Logger.WarnWithCtx(ctx, "Failed to get authorization token", "error", err.Error())
		return nil, errors.Wrap(err, "Failed to get authorization token from ecr client")
	}
	r.Logger.DebugWithCtx(ctx, "Got authorization token response")

	// AuthorizationData is a list as it used to return a token per registry, that is now deprecated.
	// The returned token can be used to access any Amazon ECR registry that the IAM principal has access to.
	for _, auth := range resp.AuthorizationData {
		token := &registry.Token{
			SecretName:  r.SecretName,
			Namespace:   r.Namespace,
			Auth:        *auth.AuthorizationToken,
			RegistryUri: r.RegistryUri,
		}
		r.Logger.InfoWithCtx(ctx, "Got authorization token", "ExpiresAt", auth.ExpiresAt)
		return token, nil
	}

	return nil, errors.New("Failed to retrieve access token")
}

func (r *Registry) createECRClient(ctx context.Context) *ecr.ECR {
	r.Logger.DebugWithCtx(ctx, "Creating ECR Client",
		"region", r.awsCreds.Region,
		"assumeRole", r.awsCreds.AssumeRole)

	sessionInstance := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(r.awsCreds.Region),
		Credentials: credentials.NewStaticCredentials(r.awsCreds.AccessKeyID,
			r.awsCreds.SecretAccessKey,
			"")}))

	if r.awsCreds.AssumeRole != "" {
		creds := stscreds.NewCredentials(sessionInstance, r.awsCreds.AssumeRole)
		return ecr.New(sessionInstance, &aws.Config{Credentials: creds})
	}

	return ecr.New(sessionInstance)
}
