package registry

type Registry interface {

	// EnrichAndValidate validate the requested registry kind params and enrich extra
	EnrichAndValidate() error

	// GetAuthToken get the authorization token to
	GetAuthToken() (*Token, error)
}
