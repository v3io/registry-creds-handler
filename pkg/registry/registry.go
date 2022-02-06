package registry

import "context"

type Registry interface {

	// EnrichAndValidate the requested registry kind params
	EnrichAndValidate() error

	// GetAuthToken get an authorization token for the registry
	GetAuthToken(ctx context.Context) (*Token, error)
}
