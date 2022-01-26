package registry

import (
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type abstractRegistry struct {
	logger     logger.Logger
	SecretName string
	Namespace  string
	Creds      string
	Endpoints  []string
	Token      Token
}

func newAbstractRegistry(loggerInstance logger.Logger,
	secretName string,
	namespace string,
	creds string,
	endpoints []string,
) (*abstractRegistry, error) {
	abstractRegistry := &abstractRegistry{
		logger:     loggerInstance.GetChild("registry-generic"),
		SecretName: secretName,
		Namespace:  namespace,
		Creds:      creds,
		Endpoints:  endpoints,
	}
	return abstractRegistry, nil
}

func (ar *abstractRegistry) validateAbstractParams() error {
	if ar.SecretName == "" {
		return errors.New("Token Secret Name is required")
	}
	if ar.Namespace == "" {
		return errors.New("Namespace is required")
	}
	return nil
}
